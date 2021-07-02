package compress

import (
	"math/rand"
	"encoding/binary"
	"testing"
	"fmt"
	"bytes"

	"github.com/phil-mansfield/guppy/lib/eq"
	"github.com/phil-mansfield/guppy/lib/particles"
	"github.com/phil-mansfield/guppy/lib/snapio"
)

func TestHeader(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	hd1 := &Header{
		FixedWidthHeader{1<<8, 1<<30, 0.5, 0.27, 0.70, 100.0, 3e9},
		[]byte{5, 4, 3, 2, 1, 0}, []string{"a", "bb", "ccc", "", "eeeee"},
		[]string{"u32", "u32", "f32", "f64", "u64"},
	}
	hd2 := *hd1

	hd2.write(buf, binary.LittleEndian)

	hd3 := &Header{ }
	hd3.read(buf, binary.LittleEndian)

	hd1.Names = append(hd1.Names, "id")
	hd1.Types = append(hd1.Types, "u64")

	if hd3.FixedWidthHeader != hd1.FixedWidthHeader {
		t.Errorf("Written fixed-width header = %v, but read header = %v.",
			hd1, hd3)
	} else if !eq.Bytes(hd1.OriginalHeader, hd3.OriginalHeader) {
		t.Errorf("Written original header = %d, but read original header = %d.",
			hd1.OriginalHeader, hd3.OriginalHeader)
	} else if !eq.Strings(hd1.Names, hd3.Names) {
		t.Errorf("Written names = %s, but read names = %s.",
			hd1.Names, hd3.Names)
	} else if !eq.Strings(hd1.Types, hd3.Types) {
		t.Errorf("Written types = %s, bute read types = %s.",
			hd1.Types, hd3.Types)
	}
}

func TestFileSmall(t *testing.T) {
	span0 := [3]int{ 3, 4, 5 }
	x0 := make([]float64, span0[0]*span0[1]*span0[2])
	span1 := [3]int{ 4, 4, 4 }
	x1 := make([]float32, span1[0]*span1[1]*span1[2])
	span2 := [3]int{18, 8, 2}
	x2 := make([]uint64, span2[0]*span2[1]*span2[2])
	span3 := [3]int{8, 8, 1}
	x3 := make([]uint32, span3[0]*span3[1]*span3[2])

	for i := range x0 { x0[i] = rand.Float64() - 0.5 }
	for i := range x1 { x1[i] = float32(rand.Float64()) - 0.5 }
	for i := range x2 { x2[i] = uint64(rand.Intn(100)) }
	for i := range x3 { x3[i] = uint32(rand.Intn(100)) }

	fields := []particles.Field{
		particles.NewFloat64("x[0]", x0), particles.NewFloat32("x[1]", x1),
		particles.NewUint64("x[2]", x2), particles.NewUint32("x3", x3),
	}
	deltas := []float64{ 1e-3, 1e-3, 0, 0 }
	order := binary.LittleEndian

	methods := []Method{
		NewLagrangianDelta(span0, deltas[0]),
		NewLagrangianDelta(span1, deltas[1]),
		NewLagrangianDelta(span2, deltas[2]),
		NewLagrangianDelta(span3, deltas[3]),
	}

	buf := NewBuffer(0)
	b := []byte{ }

	fakeFile, _ := snapio.NewFakeFile(
		[]string{"x", "x3"},
		[]interface{}{[]float32{}, []float64{}}, 1000, order,
	)
	fakeHd, _ := fakeFile.ReadHeader()

	var err error
	wr := NewWriter("test_files/small_test.gup", fakeHd, buf, b, order)
	for i := range fields {
		err = wr.AddField(fields[i], methods[i])
		if err != nil { t.Fatalf("Error in AddField('%s'): %s",
			fields[i].Name(), err.Error())
		}
	}
	b, err = wr.Flush()
	if err != nil {
		t.Fatalf("Error in Flush(): %s", err.Error())
	}

	rd, err := NewReader("test_files/small_test.gup", buf, []byte{ })
	if err != nil { t.Fatalf("Error in NewReader(): %s", err.Error()) }

	names, types := rd.Names, rd.Types
	expNames := []string{"x[0]", "x[1]", "x[2]", "x3", "id"}
 	if !eq.Strings(names, expNames) {
 		t.Errorf("Expected Reader.Names to give %s, got %s.", expNames, names)
 	}
	expTypes := []string{"f64", "f32", "u64", "u32", "u64"}
 	if !eq.Strings(types, expTypes) {
 		t.Errorf("Expected Reader.Names to give %s, got %s.", expNames, names)
 	}

	for i := range names {
		if names[i] == "id" { continue }

		f, err := rd.ReadField(names[i])
		if err != nil {
			t.Errorf("Error in ReadField('%s'): %s", names[i], err.Error())
			continue
		}
		if f.Name() != names[i] {
			t.Errorf("ReadField called on '%s', but the field '%s' was " + 
				"returned", f.Name(), names[i])
			continue
		}

		switch i {
		case 0:
			x, ok := f.Data().([]float64)
			if !ok { 
				t.Errorf("Expected '%s' to be type []float64.", f.Name())
			}
			if !eq.Float64sEps(x, x0, 1e-3) {
				t.Errorf("Expected '%s' to be %.3f, got %.3f",
					f.Name(), x0, x)
			}
		case 1:
			x, ok := f.Data().([]float32)
			if !ok { 
				t.Errorf("Expected '%s' to be type []float32.", f.Name())
			}
			if !eq.Float32sEps(x, x1, 1e-3) {
				t.Errorf("Expected '%s' to be %.3f, got %.3f",
					f.Name(), x1, x)
			}
		case 2, 3:
			if !eq.Generic(f.Data(), fields[i].Data()) {
				t.Errorf("Expected '%s' to be %v, got %v.",
					f.Name(), f.Data(), fields[i].Data())
			}
		}
	}

	rd.Close()
}

func TestFileLarge(t *testing.T) {
	// File information
	span := [3]int{ 128, 128, 128 }
	fileName := "../../large_test_data/L125_sheet125_snap_100.gadget2.dat"
	types := []string{"v32", "v32", "u32"}
	names := []string{"x", "v", "id"}
	order := binary.LittleEndian

	// Read x and v
	f, err := snapio.NewGadget2Cosmological(fileName, names, types, order)
	if err != nil { t.Fatalf(err.Error()) }
	x, v := quickRead(f, t)

	xDelta := 2.4e-3
	vDelta := 1.0

	buf := NewBuffer(0)
	b := []byte{ }

	snapHd, err := f.ReadHeader()
	if err != nil { t.Fatalf(err.Error()) }

	wr := NewWriter("test_files/large_test.gup", snapHd, buf, b, order)
	xMethod := NewLagrangianDelta(span, xDelta)
	vMethod := NewLagrangianDelta(span, vDelta)

	for k := 0; k < 3; k++ {
		xx := make([]float32, len(x))
		vv := make([]float32, len(v))

		for j := range xx { xx[j] = x[j][k] }
		for j := range vv { vv[j] = v[j][k] }

		xComp := particles.NewFloat32(fmt.Sprintf("x[%d]", k), xx)
		vComp := particles.NewFloat32(fmt.Sprintf("v[%d]", k), vv)

		err = wr.AddField(xComp, xMethod)
		if err != nil { t.Fatalf("Error in AddField('%s'): %s",
			xComp.Name(), err.Error())
		}

		err = wr.AddField(vComp, vMethod)
		if err != nil { t.Fatalf("Error in AddField('%s'): %s",
			vComp.Name(), err.Error())
		}
	}

	b, err= wr.Flush()
	if err != nil { t.Fatalf(err.Error()) }

	rd, err := NewReader("test_files/large_test.gup", buf, []byte{ })
	if err != nil { t.Fatalf("Error in NewReader(): %s", err.Error()) }

	names = rd.Names
	expNames := []string{"x[0]", "v[0]", "x[1]", "v[1]", "x[2]", "v[2]", "id"}
	dims := []int{ 0, 0, 1, 1, 2, 2 }
 	if !eq.Strings(names, expNames) {
 		t.Errorf("Expected Reader.Names to give %s, got %s.", expNames, names)
 	}

 	for i := range names {
 		if names[i] == "id" { continue }
 
 		field, err := rd.ReadField(names[i])
 		if err != nil { t.Fatalf(err.Error()) }

 		data, ok := field.Data().([]float32)
 		if !ok { t.Fatalf("Expected field to have type []float32.") }

 		dataExp := make([]float32, len(data))
 		delta, target := xDelta, x
 		if field.Name()[0] == 'v' { delta, target = vDelta, v }
 		for j := range dataExp { dataExp[j] = target[j][dims[i]] }

 		if !verboseFloat32sEps(data, dataExp, float32(delta)) {
 			t.Errorf("Field '%s' not compressed to the correct accuracy: " + 
 				"expected %.4f..., got %.4f...",
 				field.Name(), dataExp[:3], data[:3])
 		}
 	}
 }

func verboseFloat32sEps(x, y []float32, eps float32) bool {
	if len(x) != len(y) {
		fmt.Printf("len(x) = %d, len(y) = %d.\n", len(x), len(y))
		return false
	}
	for i := range x {
		if x[i] + eps < y[i] || y[i] + eps < x[i] {
			fmt.Printf("Failure at i = %d, x = %.4f, y = %.4f, delta = %.4f.\n",
				i, x[i], y[i], x[i] - y[i])
			return false
		}
	}
	return true
}

func quickRead(f snapio.File, t *testing.T) (x, v [][3]float32) {
	hd, err :=  f.ReadHeader()
	if err != nil { t.Fatalf(err.Error()) }
	buf, err := snapio.NewBuffer(hd)
	if err != nil { t.Fatalf(err.Error()) }

	err = f.Read("x", buf)
	if err != nil { t.Fatalf(err.Error()) }
	err = f.Read("v", buf)
	if err != nil { t.Fatalf(err.Error()) }

	xIntr, err := buf.Get("x")
	if err != nil { t.Fatalf(err.Error()) }
	vIntr, err := buf.Get("v")
	if err != nil { t.Fatalf(err.Error()) }

	var ok bool
	x, ok = xIntr.([][3]float32)
	if !ok { t.Fatalf("x is not [][3]float32") }
	v, ok = vIntr.([][3]float32)
	if !ok { t.Fatalf("v is not [][3]float32") }

	return x, v
}
