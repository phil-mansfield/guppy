package main

import (
	"fmt"
	"encoding/binary"
	"github.com/phil-mansfield/guppy/lib/snapio"
)

const (
	SimName = "Erebos_CBol_L63"
)

func InName(snap, i int) string {
    return fmt.Sprintf("/data/mansfield/simulations/%s/" +
        "particles/raw/snapdir_%03d/snapshot_%03d.%d",
        SimName, snap, snap, i)
}

func InName2(snap, i int) string {
    return fmt.Sprintf("/data/mansfield/simulations/%s/" +
        "particles/gadget_cube/snapdir_%03d/sheet%d%d%d.dat",
        SimName, snap, i, i, i)
}

func main() {
	fmt.Println("# snapshot, z, a")
	for snap := 0; snap <=100; snap++ {
		if snap == 63 {
			fmt.Printf(" 63 -1      -1\n")
			continue
		}
		name := InName2(snap, 0)
		file, err := snapio.NewGadget2Cosmological(name,
			[]string{"x", "v", "id"}, []string{"v32", "v32", "u32"},
			binary.LittleEndian)
		if err != nil { panic(err.Error()) }

		hd, err := file.ReadHeader()
		if err != nil { panic(err.Error()) }

		z := hd.Z()
		a := 1/(1 + z)
		fmt.Printf("%3d %.6f %.6f\n", snap, z, a)
	}
}
