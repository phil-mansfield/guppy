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
)

var (
	XDelta = 0.005
	VDelta = 5.0

	SimPath = "/data/mansfield/simulations/Erebos_CBol_L63_N256/particles/raw"
	GuppyPath = "/data/mansfield/simulations/Erebos_CBol_L63_N256/particles/guppy/dx5_dv5"
	SnapshotFormat = "snapdir_%04d"
	InputFileFormat = "snapshot_%04d.%d"
	OutputFileFormat = "snapshot_%04d.%d.gup"

	NFiles = 16
	Np = 256
	NBlocks = 2
	Snaps = []int{189, 190}

	GadgetVarNames = []string{"x", "v", "id"}
	GadgetVarTypes = []string{"v32", "v32", "u64"} 

	Order = binary.LittleEndian
	Workers = 4
)

func GadgetFileName(snap, i int) string {
	format := path.Join(SimPath, SnapshotFormat, InputFileFormat)
	return fmt.Sprintf(format, snap, snap, i)
}

func GuppyFileName(snap, i int) string {
	format := path.Join(GuppyPath, SnapshotFormat, OutputFileFormat)
	return fmt.Sprintf(format, snap, snap, i)
}

func OutputDir(snap int) string {
	format := path.Join(GuppyPath, SnapshotFormat)
	return fmt.Sprintf(format, snap)
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
	id, ok = idIntr.([]uint64)
	if !ok { panic("Type error on id") }

	return x, v, id
}

func LoadGadgetFile(
	snap, i int, xGrid, vGrid [][3]float32, sioBuf *snapio.Buffer,
) {
	x, v, id := GetGadgetArrays(snap, i, sioBuf)

	Np64 := uint64(Np)

	for i := range x {
		ix := int(id[i] / (Np64*Np64)) 
		iy := int((id[i] / Np64) % Np64)
		iz := int(id[i] % Np64)

		j := ix + Np*iy + Np*Np*iz

		xGrid[j] = x[i]
		vGrid[j] = v[i]
	}

	sioBuf.Reset()

}

func BaseHeader() snapio.Header {
	baseFile := GadgetFileName(Snaps[len(Snaps) - 1], 0)
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

func OutputBuffers(Nb int) *Output {
	span := [3]int{ Nb, Nb, Nb }
	methodMap := map[string]compress.Method{
		"x[0]": compress.NewLagrangianDelta(span, XDelta),
		"x[1]": compress.NewLagrangianDelta(span, XDelta),
		"x[2]": compress.NewLagrangianDelta(span, XDelta),
		"v[0]": compress.NewLagrangianDelta(span, VDelta),
		"v[1]": compress.NewLagrangianDelta(span, VDelta),
		"v[2]": compress.NewLagrangianDelta(span, VDelta),
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
		iz := jz + offsetZ
		for jy := 0; jy < Nb; jy++ {
			iy := jy + offsetY
			for jx := 0; jx < Nb; jx++ {
				ix := jx + offsetX

				iBlock := jx + jy*Nb + jz*Nb*Nb
				iGrid := ix + iy*Nb + iz*Nb*Nb

				output.Data[iBlock] = grid[iGrid][dim]
			}
		}
	}

	componentName := fmt.Sprintf("%s[%d]", varName, dim)
	component := particles.NewFloat32(componentName, output.Data)

	err := output.Writer.AddField(component, output.Methods[componentName])
	if err != nil { panic(err.Error()) }
}

func WriteToGuppy(
	snap, i int, xGrid, vGrid [][3]float32,
	origHd snapio.Header, output *Output,
) {
	outName := GuppyFileName(snap, i)

	Nb := Np / NBlocks		
	offsetID := uint64(Nb*Nb*Nb * i)
	span := [3]int64{ int64(Nb), int64(Nb), int64(Nb) }

	
	output.Writer = compress.NewWriter(
		outName, origHd, offsetID, span, output.Buffer, output.B, Order,
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

	xGrid := make([][3]float32, Np*Np*Np)
	vGrid := make([][3]float32, Np*Np*Np)
	baseHd := BaseHeader()

	Nb := Np / NBlocks

	outBufs := make([]*Output, Workers)
	sioBufs := make([]*snapio.Buffer, Workers)
	for i := range sioBufs {
		var err error
		sioBufs[i], err = snapio.NewBuffer(baseHd)
		if err != nil { panic(err.Error()) }
		outBufs[i] = OutputBuffers(Nb)
	}

	log.Println("Finished setup.")
	
	for _, snap := range Snaps {
		log.Println("Analyzing snap", snap)

		log.Println("Running SplitArray(LoadGadgetFile)")
		thread.SplitArray(NFiles, Workers, func(w, start, end, step int) {
			sioBuf := sioBufs[w]
			for i := start; i < end; i += step {
				LoadGadgetFile(snap, i, xGrid, vGrid, sioBuf)
			}
		})
		
		err := os.Mkdir(OutputDir(snap), 0755) 
		if err != nil && !strings.Contains(err.Error(), "file exists") {
			panic(err.Error())
		}
		
		log.Println("Running SplitArray(WriteToGuppy)")
		nGuppyFiles := NBlocks*NBlocks*NBlocks
		thread.SplitArray(nGuppyFiles, Workers, func(w, start, end, step int) {
			output := outBufs[w]
			for i := start; i < end; i++ {
				WriteToGuppy(snap, i, xGrid, vGrid, baseHd, output)
			}
		})
	}
}
