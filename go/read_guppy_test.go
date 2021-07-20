package read_guppy

import (
	"fmt"
	"testing"
)

func TestEverything(t *testing.T) {
	fname := "../large_test_data/large_test.gup"

	hd := ReadHeader(fname)
	fmt.Println(hd)

	nWorkers := 2
	InitWorkers(nWorkers)

	varNames := []string{ "x[0]", "x[1]", "x[2]", "v[0]", "v[1]", "v[2]",
		"id", "x", "v" }
	varBufs := []interface{} {
		make([]float32, hd.N), make([]float32, hd.N), make([]float32, hd.N),
		make([]float32, hd.N), make([]float32, hd.N), make([]float32, hd.N),
		make([]uint64, hd.N), make([][3]float32, hd.N), make([][3]float32, hd.N),
	}

	for i := range varNames {
		workerID := i % nWorkers

		ReadVar(fname, varNames[i], workerID, varBufs[i])

		switch x := varBufs[i].(type) {
		case []float32:
			fmt.Printf("%s (f32): %.3f\n", varNames[i], x[:4])
		case [][3]float32:
			fmt.Printf("%s (f32): %.3f\n", varNames[i], x[:4])
		case []uint64:
			fmt.Printf("%s (f32): %d\n", varNames[i], x[:4])
		default: panic("Impossible")
		}
	}
}