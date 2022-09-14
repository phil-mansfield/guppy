package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"

	"github.com/phil-mansfield/guppy/lib/snapio"
	guppy "github.com/phil-mansfield/guppy/go"

	"unsafe"
)
var (
	InputPath = "/data/mansfield/simulations/Erebos_CBol_L63_N256/particles/guppy/dx1.25_dv1.25"
	OutputPath = "/data/mansfield/simulations/Erebos_CBol_L63_N256/particles/noisy_gadget/dx1.25_dv1.25_2"
	SnapshotFormat = "snapdir_%04d"
	InputFileFormat = "snapshot_%04d.%d.gup"
	OutputFileFormat = "snapshot_%04d.%d"

	BlocksOnSide = 2
	Blocks = BlocksOnSide*BlocksOnSide*BlocksOnSide
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

func ReadGuppyFile(
	snap, block int,
) (hd *guppy.Header, x, v [][3]float32, id []uint64) {
	fileName := InputFileName(snap, block)

	hd = guppy.ReadHeader(fileName)

	x = make([][3]float32, hd.N)
	v = make([][3]float32, hd.N)
	id = make([]uint64, hd.N)

	guppy.ReadVar(fileName, "x", 0, x)
	guppy.ReadVar(fileName, "v", 0, v)
	guppy.ReadVar(fileName, "id", 0, id)
	
	return hd, x, v, id
}

type gadget2Header struct {
	NPart                                     [6]uint32
    Mass                                      [6]float64
    Time, Redshift                            float64
    FlagSfr, FlagFeedback                     int32
    NPartTotal                                [6]uint32
    FlagCooling, NumFiles                     int32
    BoxSize, Omega0, OmegaLambda, HubbleParam float64
    FlagStellarAge, HashTabSize               int32
	NPartTotalHW                              [6]uint32

    Padding [64]byte
}

func WriteGadgetFile(
	snap, block int, hd *guppy.Header,
	x, v [][3]float32, id []uint64,
) {	
	fileName := OutputFileName(snap, block)

	ghd := gadget2Header{
		NPart: [6]uint32{ 0, uint32(hd.N), 0, 0, 0, 0 },
		Mass: [6]float64{ 0, hd.Mass/1e10, 0, 0, 0, 0 },
		Time: 1/(1+hd.Z), Redshift: hd.Z,
		NPartTotal: [6]uint32{ 0, uint32(hd.NTot), 0, 0, 0, 0 }, 
		NumFiles: int32(Blocks), BoxSize: hd.L, Omega0: hd.OmegaM,
		OmegaLambda: hd.OmegaL, HubbleParam: hd.H100,
	}

	f, err := os.Create(fileName)
	if err != nil { panic(err.Error()) }
	defer f.Close()

	hdSize := uint32(unsafe.Sizeof(ghd))
	if hdSize != 256 { panic("Incorrect Header Size") }
	xSize := uint32(int(unsafe.Sizeof(x[0]))*len(x))
	vSize := uint32(int(unsafe.Sizeof(v[0]))*len(v))
	idSize := uint32(int(unsafe.Sizeof(id[0]))*len(id))

	err = binary.Write(f, binary.LittleEndian, hdSize)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, ghd)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, hdSize)
	if err != nil { panic(err.Error()) }

	err = binary.Write(f, binary.LittleEndian, xSize)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, x)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, binary.LittleEndian, xSize)
	if err != nil { panic(err.Error()) }

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

func ConvertGuppyIDs(id []uint64) {
	nSide := uint64(BlocksOnSide)
	Np64 := uint64(Np)
	Nb64 := Np64 / nSide
	Nbtot64 := Nb64*Nb64*Nb64

	for i := range id {
		b := id[i] / Nbtot64
		bx, by, bz := b%nSide, (b/nSide) % nSide, b/(nSide*nSide)

		offset := b*Nbtot64
		j := id[i] - offset
		jx, jy, jz := j%Nb64, (j/Nb64) % Nb64, j / (Nb64*Nb64)
		ix, iy, iz := bx*Nb64 + jx, by*Nb64 + jy, bz*Nb64 + jz

		id[i] = 1 + iz + iy*Np64 + ix*Np64*Np64
	}
}

func main() {
	guppy.InitWorkers(1)

	for _, snap := range Snaps {
		for block := 0; block < Blocks; block++ {
			fmt.Println(snap, block)
			hd, x, v, id := ReadGuppyFile(snap, block)
			ConvertGuppyIDs(id)
			WriteGadgetFile(snap, block, hd, x, v, id)
		}
	}
}
