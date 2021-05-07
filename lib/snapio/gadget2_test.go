package snapio

import (
	"encoding/binary"
	"testing"
	"math"
	"unsafe"

	"github.com/phil-mansfield/guppy/lib/eq"
)

var (
	fileName = "../../large_test_data/snapshot_023.39"
	names = []string{"x", "v", "id"}
	types = []string{"v32", "v32", "u64"}
	order = binary.LittleEndian
)


func TestGadget2HeaderSize(t *testing.T) {
	if size := unsafe.Sizeof(rawGadget2Header{ }); size != 256 {
		t.Errorf("rawGadget2Header{} has size %d, not 256", size)
	}
	if size := unsafe.Sizeof(rawLGadget2Header{ }); size != 256 {
		t.Errorf("rawLGadget2Header{} has size %d, not 256", size)
	}
}
func TestGadget2CosmologicalFailure(t *testing.T) {
	fileNames := []string{"file_that_doesn't_exist.dat", "test_files",
		"test_files/tiny_file.txt"}

	for _, fileName := range fileNames {
		_, err := NewGadget2Cosmological(fileName, names, types, order)
		if err == nil { 
			t.Errorf("Expected read of %s to fail, but succeeded.", fileName)
		}
	}

	tests := []struct{
		names, types []string
	} {
		{ []string{}, []string{} },
		{ []string{"x"}, []string{"x32"} },
		{ []string{"x"}, []string{"f32"} },
		{ []string{"v"}, []string{"f32"} },
		{ []string{"id"}, []string{"v32"} },
		{ []string{"x", "x", "id"}, []string{"v32", "v32", "u64"} },
		{ []string{"x", "v", "id"}, []string{"v32", "v32", "u32"} },
		{ []string{"x", "v", "id", "phi"},
			[]string{"v32", "v32", "u32", "f32"} },
	}

	for i := range tests {
		_, err := NewGadget2Cosmological(
			fileName, tests[i].names, tests[i].types, order,
		)
		if err == nil {
			t.Errorf("Expected read with names = %s and types = %s to fail, " + 
				"but succeeded", tests[i].names, tests[i].types)
		}
	}
}

func TestReadGadget2Header(t *testing.T) {
	f, err := NewGadget2Cosmological(fileName, names, types, order)
	if err != nil {
		t.Fatalf("Expected valid read, got error message %s.", err.Error())
	}

	hd, err := f.ReadHeader()
	if err != nil {
		t.Fatalf("Expected valid header read, got error message %s.",
			err.Error())
	}

	nExp := []string{"x", "v", "id"}
	if n := hd.Names(); !eq.Strings(n, nExp) {
		t.Errorf("Expected hd.Names() = %s, got %s", nExp, n)
	}

	tyExp := []string{"x", "v", "id"}
	if ty := hd.Names(); !eq.Strings(ty, tyExp) {
		t.Errorf("Expected hd.Names() = %s, got %s", tyExp, ty)
	}

	nTotExp := int64(1)<<30
	if nTot := hd.NTot(); nTot != nTotExp {
		t.Errorf("Expected hd.NTot() = %d, got %d.", nTotExp, nTot)
	}

	zExp := 13.917467933963309
	if z := hd.Z(); z != zExp {
		t.Errorf("Expected hd.z = %f, got %f.", zExp, z)
	}

	omegaMExp := 0.286
	if omegaM := hd.OmegaM(); omegaM != omegaM {
		t.Errorf("Expected hd.omegaM = %f, got %f.", omegaMExp, omegaM)
	}

	h100Exp := 0.7
	if h100 := hd.H100(); h100 != h100Exp {
		t.Errorf("Expected hd.h100 = %f, got %f.", h100Exp, h100)
	}

	LExp := 125.0
	if L := hd.L(); L != LExp {
		t.Errorf("Expected hd.z = %f, got %f.", LExp, L)
	}

	mpExp := 1.4439009231682864e+08
	if mp := hd.Mass(); mp != mpExp {
		t.Errorf("Expected hd.Maa = %g, got %g.", mpExp, mp)
	}
}

func TestReadGadget2Data(t *testing.T) {
	f, err := NewGadget2Cosmological(fileName, names, types, order)
	if err != nil {
		t.Fatalf("Expected valid read, got error message %s.", err.Error())
	}

	hd, err := f.ReadHeader()
	if err != nil {
		t.Fatalf("Expected valid header read, got error message %s.",
			err.Error())
	}

	buf, err := NewBuffer(hd)
	if err != nil {
		t.Fatalf("Expected Buffer could be created, got error message %s.",
			err.Error())
	}

	err = f.Read("x", buf)
	if err != nil { t.Fatalf("Got error '%s' when reading x", err.Error()) }
	err = f.Read("v", buf)
	if err != nil { t.Fatalf("Got error '%s' when reading v", err.Error()) }
	err = f.Read("id", buf)
	if err != nil { t.Fatalf("Got error '%s' when reading id", err.Error()) }

	xIntr, err := buf.Get("x")
	if err != nil { t.Errorf("Couldn't Get() x: %s", err.Error()) }
	vIntr, err := buf.Get("v")
	if err != nil { t.Errorf("Couldn't Get() v: %s", err.Error()) }
	idIntr, err := buf.Get("id")
	if err != nil { t.Errorf("Couldn't Get() id: %s", err.Error()) }

	x, ok := xIntr.([][3]float32)
	if !ok { t.Fatalf("Incorrect type for buf.Get('x')") }
	v, ok := vIntr.([][3]float32)
	if !ok { t.Fatalf("Incorrect type for buf.Get('v')") }
	id, ok := idIntr.([]uint64)
	if !ok { t.Fatalf("Incorrect type for buf.Get('id')") }

	xExp := [][3]float32{
		{101.70, 93.76, 54.62},
		{101.69, 93.74, 54.52},
		{101.70, 93.62, 54.49},
		{101.71, 93.65, 54.60},
		{101.81, 93.65, 54.61},
	}
	vExp := [][3]float32{
		{113.47, 3.75, 41.97},
		{108.05, -7.03, 58.02},
		{115.16, -7.33, 39.29},
		{119.56, 10.11, 30.37},
		{106.54, 13.20, 37.18},
	}
	idExp := []uint64{873202112, 873202111, 873201087, 873201088, 874249664}

	if !vecsAlmostEq(x[:5], xExp) {
		t.Errorf("Expected first five position vectors to be %.2f, got %.2f",
			x[:5], xExp)
	} 
	if !vecsAlmostEq(v[:5], vExp) {
		t.Errorf("Expected first five velcotiy vectors to be %.2f, got %.2f",
			v[:5], vExp)
	}
	if !eq.Uint64s(id[:5], idExp) {
		t.Errorf("Expected first five IDs to be %d, got %d.", id[:5], idExp)
	}
}

func almostEq(x, y float64) bool {
	eps := 1e-3
	return math.Abs(x - y) < eps*math.Abs(x)
}

func vecAlmostEq(x, y [3]float32) bool {
	for k := 0; k < 3; k++ {
		if !almostEq(float64(x[k]), float64(y[k])) {
			return false
		}
	}
	return true
}

func vecsAlmostEq(x, y [][3]float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if !vecAlmostEq(x[i], y[i]) { return false }
	}
	return true
}