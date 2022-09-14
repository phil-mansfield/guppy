package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"

	"github.com/phil-mansfield/guppy/lib/snapio"

	"unsafe"
)

var (
	XDelta = 0.001
	VDelta = 1.00
	L = 62.5

	InputPath = "/data/mansfield/simulations/Erebos_CBol_L63_N256/particles/raw"
	OutputPath = "/data/mansfield/simulations/Erebos_CBol_L63_N256/particles/noisy_gadget/dx1_dv1"
	SnapshotFormat = "snapdir_%04d"
	InputFileFormat = "snapshot_%04d.%d"
	OutputFileFormat = "snapshot_%04d.%d"

	Blocks = 16
	Np = 256
	Snaps = []int{189, 190}

	GadgetVarNames = []string{"x", "v", "id"}
	GadgetVarTypes = []string{"v32", "v32", "u64"} 
	Order = binary.LittleEndian
)

func InputFileName(snap, block int) string {
	format := path.Join(InputPath, SnapshotFormat, InputFileFormat)
	return fmt.Sprintf(format, snap, snap, block)
}

func OutputFileName(snap, block int) string {
	format := path.Join(OutputPath, SnapshotFormat, OutputFileFormat)
	return fmt.Sprintf(format, snap, snap, block)
}

func BaseHeader() snapio.Header {
	baseFile := InputFileName(Snaps[len(Snaps) - 1], 0)
	reader, err := snapio.NewLGadget2(
		baseFile, GadgetVarNames, GadgetVarTypes, Order,
	)
	if err != nil { panic(err.Error()) }

	hd, err := reader.ReadHeader()
	if err != nil { panic(err.Error()) }
	return hd
}


func GetGadgetArrays(
	snap, i int, sioBuf *snapio.Buffer,
) (hd snapio.Header, x, v [][3]float32, id []uint64) {
	fileName := InputFileName(snap, i)
	reader, err := snapio.NewLGadget2(
		fileName, GadgetVarNames, GadgetVarTypes, Order,
	)
	if err != nil { panic(err.Error()) }
	
	hd, err = reader.ReadHeader()
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

	sioBuf.Reset()

	return hd, x, v, id
}
func PeriodicDx(dx, L float64) float64 {
	pix :=math.Ceil(L / dx)
	return L / pix
}

func Noise(x [][3]float32, dx float64) {
	for i := range x {
		for dim := 0; dim < 3; dim++ {
			bin := math.Floor(float64(x[i][dim]) / dx)
			x[i][dim] = float32((bin + rand.Float64()) * dx)
		}
	}
}


func WriteGadgetFile(
	snap, block int, hd snapio.Header,
	x, v [][3]float32, id []uint64,
) {	
	fileName := OutputFileName(snap, block)

	f, err := os.Create(fileName)
	if err != nil { panic(err.Error()) }
	defer f.Close()
	
	ohd := hd.ToBytes()

	hdSize := uint32(len(ohd))
	if hdSize != 256 { panic("Incorrect Header Size") }
	xSize := uint32(int(unsafe.Sizeof(x[0]))*len(x))
	vSize := uint32(int(unsafe.Sizeof(v[0]))*len(v))
	idSize := uint32(int(unsafe.Sizeof(id[0]))*len(v))

	err = binary.Write(f, binary.LittleEndian, hdSize)
	if err != nil { panic(err.Error()) }
	_, err = f.Write(ohd)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, hdSize)
	if err != nil { panic(err.Error()) }

	err = binary.Write(f, binary.LittleEndian, xSize)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, x)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, xSize)
	if err != nil { panic(err.Error()) }

	/*
	rootA := float32(math.Sqrt(1 / (1 + hd.Z())))
	for i := range v {
		for dim := 0; dim < 3; dim++ {
			v[i][dim] /= rootA
		}
	}
    */

	err = binary.Write(f, binary.LittleEndian, vSize)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, v)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, vSize)
	if err != nil { panic(err.Error()) }

	err = binary.Write(f, binary.LittleEndian, idSize)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, id)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, idSize)
	if err != nil { panic(err.Error()) }
}


func main() {
	baseHd := BaseHeader()
	sioBuf, err := snapio.NewBuffer(baseHd)
	if err != nil { panic(err.Error()) }

	for _, snap := range Snaps {
		for block := 0; block < Blocks; block++ {
			fmt.Println(snap, block)
			hd, x, v, id := GetGadgetArrays(snap, block, sioBuf)

			Noise(x, PeriodicDx(XDelta, L))
			Noise(v, VDelta)

			WriteGadgetFile(snap, block, hd, x, v, id)
		}
	}
}
