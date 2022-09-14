package main

import (
	"fmt"
	"os"
	"path"
	"log"
	"strings"
	"encoding/binary"
	"github.com/phil-mansfield/guppy/lib/compress"
	"github.com/phil-mansfield/guppy/lib/particles"
	"github.com/phil-mansfield/guppy/lib/snapio"
	"github.com/phil-mansfield/guppy/lib/thread"
	"runtime"
)

var (
	XDelta = 0.001
	VDelta = 17.20/8

	MultNames = []string{"d8", "x_d2", "x", "x_m2", "x_m4",
		"v_d2", "v", "v_m2", "v_m4", "v_m8"}
	XMults = []float64{0.125,   0.5,   1.0,   2.0,   4.0,
		0.125, 0.125, 0.125, 0.125, 0.125}
	VMults = []float64{0.125, 0.125, 0.125, 0.125, 0.125,
		  0.5,   1.0,   2.0,   4.0,   8.0}

	SimPath = "/data/mansfield/simulations/Erebos_CBol_L63/particles/raw"
	GuppyPathFmt = "/data/mansfield/simulations/Erebos_CBol_L63/particles/guppy/fid_%s"
	SnapshotFormat = "snapdir_%03d"
	InputFileFormat = "snapshot_%03d.%d"
	OutputFileFormat = "snapshot_%03d.%d.gup"

	NFiles = 512
	Np = 1024
	NBlocks = 8
	Snaps = []int{100}

	GadgetVarNames = []string{"x", "v", "id"}
	GadgetVarTypes = []string{"v32", "v32", "u64"} 
	
	Order = binary.LittleEndian
	Workers = 8
)

func GadgetFileName(snap, i int) string {
	format := path.Join(SimPath, SnapshotFormat, InputFileFormat)
	return fmt.Sprintf(format, snap, snap, i)
}

func GuppyFileName(multName string, snap, i int) string {
	format := path.Join(GuppyPathFmt, SnapshotFormat, OutputFileFormat)
	return fmt.Sprintf(format, multName, snap, snap, i)
}

func OutputDir(multName string, snap int) string {
	format := path.Join(GuppyPathFmt, SnapshotFormat)
	return fmt.Sprintf(format, multName, snap)
}

func GetGadgetArrays(
	snap, i int, sioBuf *snapio.Buffer,
) (x, v [][3]float32, id []uint64) {
	fileName := GadgetFileName(snap, i)
	reader, err := snapio.NewLGadget2(
		fileName, GadgetVarNames, GadgetVarTypes, Order,
	)
	if err != nil { panic(err.Error()) }

	err = reader.Read("x", sioBuf)
	if err != nil { panic(err.Error()) }
	err = reader.Read("v", sioBuf)
	if err != nil { panic(err.Error()) }
	err = reader.Read("id", sioBuf)
	if err != nil { panic(err.Error()) }

	xIntr, err := sioBuf.Get("x")
	if err != nil { panic(err.Error()) }
	vIntr, err := sioBuf.Get("v")
	if err != nil { panic(err.Error()) }
	idIntr, err := sioBuf.Get("id")
	if err != nil { panic(err.Error()) }

	x, ok := xIntr.([][3]float32)
	if !ok { panic("Type error on x") }
	v, ok = vIntr.([][3]float32)
	if !ok { panic("Type error on v") }
	if id32, ok := idIntr.([]uint32); ok {
		runtime.GC()
		id = make([]uint64, len(id32))
		for i := range id {
			id[i] = uint64(id32[i])
		}
	} else { 
		id, ok = idIntr.([]uint64)
		if !ok { panic("Type error on id") }
	}

	return x, v, id
}

func LoadGadgetFile(
	snap, i int, xGrid, vGrid [][3]float32, sioBuf *snapio.Buffer,
) {
	x, v, id := GetGadgetArrays(snap, i, sioBuf)

	Np64 := uint64(Np)

	for i := range x {
		ID := id[i] - 1
		ix := int(ID / (Np64*Np64)) 
		iy := int((ID / Np64) % Np64)
		iz := int(ID % Np64)

		j := ix + Np*iy + Np*Np*iz

		xGrid[j] = x[i]
		vGrid[j] = v[i]
	}

	sioBuf.Reset()

}

func BaseHeader(snap int) snapio.Header {
	baseFile := GadgetFileName(snap, 0)
	reader, err := snapio.NewLGadget2(
		baseFile, GadgetVarNames, GadgetVarTypes, Order,
	)
	if err != nil { panic(err.Error()) }

	hd, err := reader.ReadHeader()
	if err != nil { panic(err.Error()) }
	return hd
}

type Output struct {
	Buffer *compress.Buffer
	Methods map[string]compress.Method
	Data []float32
	B []byte
	Writer *compress.Writer
}

func OutputBuffers(Nb int, xMult, vMult, L float64) *Output {
	span := [3]int{ Nb, Nb, Nb }
	methodMap := map[string]compress.Method{
		"x{0}": compress.NewLagrangianDelta(span, XDelta*xMult, L),
		"x{1}": compress.NewLagrangianDelta(span, XDelta*xMult, L),
		"x{2}": compress.NewLagrangianDelta(span, XDelta*xMult, L),
		"v{0}": compress.NewLagrangianDelta(span, VDelta*vMult, 0),
		"v{1}": compress.NewLagrangianDelta(span, VDelta*vMult, 0),
		"v{2}": compress.NewLagrangianDelta(span, VDelta*vMult, 0),
	}

	return &Output{
		compress.NewBuffer(0), methodMap,
		make([]float32, Nb*Nb*Nb),
		[]byte{ }, nil,
	}
}

func WriteField(
	varName string, grid [][3]float32, dim, i, Nb int, output *Output,
) {
	offsetZ := i / (NBlocks * NBlocks)
	offsetY := (i / NBlocks) % NBlocks
	offsetX := i % NBlocks

	for jz := 0; jz < Nb; jz++ {
		iz := jz + offsetZ*Nb
		for jy := 0; jy < Nb; jy++ {
			iy := jy + offsetY*Nb
			for jx := 0; jx < Nb; jx++ {
				ix := jx + offsetX*Nb

				iBlock := jx + jy*Nb + jz*Nb*Nb
				iGrid := ix + iy*Np + iz*Np*Np

				output.Data[iBlock] = grid[iGrid][dim]
			}
		}
	}

	componentName := fmt.Sprintf("%s{%d}", varName, dim)
	component := particles.NewFloat32(componentName, output.Data)

	err := output.Writer.AddField(component, output.Methods[componentName])
	if err != nil { panic(err.Error()) }
}

func WriteToGuppy(
	snap, i int, xGrid, vGrid [][3]float32,
	origHd snapio.Header, output *Output,
	multName string,
) {
	outName := GuppyFileName(multName, snap, i)

	Nb := Np / NBlocks		
	span := [3]int64{ int64(Nb), int64(Nb), int64(Nb) }
	bx := i % NBlocks
	by := (i / NBlocks) % NBlocks
	bz := i / (NBlocks * NBlocks)
	offset := [3]int64{ int64(Nb*bx), int64(Nb*by), int64(Nb*bz) }
	totSpan := [3]int64{ int64(Np), int64(Np), int64(Np) }
	
	output.Writer = compress.NewWriter(
		outName, origHd, span, offset, totSpan,
		output.Buffer, output.B, Order,
	)

	for dim := 0; dim < 3; dim++ {
		WriteField("x", xGrid, dim, i, Nb, output)
	}

	for dim := 0; dim < 3; dim++ {
		WriteField("v", vGrid, dim, i, Nb, output)
	}
	var err error
	output.B, err = output.Writer.Flush()
	if err != nil { panic(err.Error()) }
}

func main() {
	thread.Set(Workers)

	if len(Snaps) == 0 {
		Snaps = make([]int, 101)
		for i := range Snaps { Snaps[i] = i }
	}

	xGrid := make([][3]float32, Np*Np*Np)
	vGrid := make([][3]float32, Np*Np*Np)

	log.Println("Finished setup.")
	
	for i := range XMults {
		CompressSimulation(xGrid, vGrid, XMults[i], VMults[i], MultNames[i])
	}
}

func CompressSimulation(
	xGrid, vGrid [][3]float32,
	xMult, vMult float64, multName string,
) {
	baseHd := BaseHeader(Snaps[0])
	Nb := Np / NBlocks

	outBufs := make([]*Output, Workers)
	sioBufs := make([]*snapio.Buffer, Workers)
	for i := range sioBufs {
		var err error
		sioBufs[i], err = snapio.NewBuffer(baseHd)
		if err != nil { panic(err.Error()) }
		outBufs[i] = OutputBuffers(Nb, xMult, vMult, baseHd.L())
	}

	for _, snap := range Snaps {
		log.Println("Analyzing snap", snap)

		ghd := BaseHeader(snap)

		log.Println("Running SplitArray(LoadGadgetFile)")
		thread.SplitArray(NFiles, Workers, func(w, start, end, step int) {
			sioBuf := sioBufs[w]
			for i := start; i < end; i += step {
				LoadGadgetFile(snap, i, xGrid, vGrid, sioBuf)
			}
		})

		err := os.MkdirAll(OutputDir(multName, snap), 0755) 
		if err != nil && !strings.Contains(err.Error(), "file exists") {
			panic(err.Error())
		}
		
		log.Println("Running SplitArray(WriteToGuppy)")
		nGuppyFiles := NBlocks*NBlocks*NBlocks
		thread.SplitArray(nGuppyFiles, Workers, func(w, start, end, step int) {
			output := outBufs[w]
			for i := start; i < end; i++ {
				WriteToGuppy(snap, i, xGrid, vGrid,
					ghd, output, multName)
			}
		})
	}
}
