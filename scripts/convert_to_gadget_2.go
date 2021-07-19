package main

import (
	"encoding/binary"
	"os"
	"fmt"
	
	"github.com/phil-mansfield/gotetra/render/io"
	"github.com/phil-mansfield/gotetra/render/geom"

	"unsafe"
)

const (
	GtetFmt = "/data/mansfield/simulations/Erebos_CBol_L63/particles/gtet/snapdir_%03d/sheet%d%d%d.dat"
	GadgetCubeFmt = "/data/mansfield/simulations/Erebos_CBol_L63/particles/gadget_cube/snapdir_%03d/sheet%d%d%d.dat"

	SnapMin = 0
	SnapMax = 99
	SkipMod = -1
)
var(
	CoordMin = [3]int{5, 6, 0}
	CoordMax = [3]int{6, 7, 1}
)

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

func main() {
	for snap := SnapMin; snap <= SnapMax; snap++ {
		if snap == 63 { continue }
		if SkipMod > 0 && snap % SkipMod != 0 { continue }

		for iz := CoordMin[2]; iz <= CoordMax[2]; iz++ {
			for iy := CoordMin[1]; iy <= CoordMax[1]; iy++ {
				for ix := CoordMin[0]; ix <= CoordMax[0]; ix++ {
					Convert(snap, ix, iy, iz)
				}
			}
		}
	}
}

func Convert(snap, jx, jy, jz int) {
	inFname := fmt.Sprintf(GtetFmt, snap, jx, jy, jz)
	outFname := fmt.Sprintf(GadgetCubeFmt, snap, jx, jy, jz)
	zoff, yoff, xoff := 128*jx, 128*jy, 128*jz

	hd := &io.SheetHeader{ }
	err := io.ReadSheetHeaderAt(inFname, hd)
	if err != nil { panic(err.Error()) }
	
	gw, sw := int(hd.GridWidth), int(hd.SegmentWidth)
	xg, vg := make([]geom.Vec, gw*gw*gw), make([]geom.Vec, gw*gw*gw)
	err = io.ReadSheetPositionsAt(inFname, xg)
	if err != nil { panic(err.Error()) }
	err = io.ReadSheetVelocitiesAt(inFname, vg)
	if err != nil { panic(err.Error()) }

	x, v := make([][3]float32, sw*sw*sw), make([][3]float32, sw*sw*sw)
	id := make([]uint32, sw*sw*sw)
	for ix := 0; ix < sw; ix++ {
		for iy := 0; iy < sw; iy++ {
			for iz := 0; iz < sw; iz++ {
				ig := iz + iy*gw + ix*gw*gw
				i := iz + iy*sw + ix*sw*sw
				x[i] = [3]float32{ xg[ig][0], xg[ig][1], xg[ig][2] }
				v[i] = [3]float32{ vg[ig][0], vg[ig][1], vg[ig][2] }
				id[i] = uint32((iz+zoff) + (iy+yoff)*sw + (ix+xoff)*sw*sw)
			}
		}
	}

	ghd := &gadget2Header{
		NPart: [6]uint32{ 0, uint32(sw*sw*sw), 0, 0, 0, 0},
		NPartTotal: [6]uint32{ 0, 1<<30 , 0, 0, 0, 0},
		Mass: [6]float64{ 0, hd.Mass, 0, 0, 0, 0 },
		BoxSize: hd.TotalWidth,
		Time: 1/(1 + hd.Cosmo.Z),
		Redshift: hd.Cosmo.Z,
		Omega0: hd.Cosmo.OmegaM,
		OmegaLambda: hd.Cosmo.OmegaL,
		HubbleParam: hd.Cosmo.H100,
		NumFiles: 1,
	}
	
	f, err := os.Create(outFname)
	if err != nil { panic(err.Error()) }
	defer f.Close()
	
	hdSize := uint32(unsafe.Sizeof(*ghd))
	if hdSize != 256 { panic("Incorrect Header Size") }
	xSize := uint32(int(unsafe.Sizeof(x[0]))*len(x))
	vSize := uint32(int(unsafe.Sizeof(v[0]))*len(v))
	idSize := uint32(int(unsafe.Sizeof(id[0]))*len(v))

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
