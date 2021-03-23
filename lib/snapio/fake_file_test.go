package snapio

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestFakeFile(t *testing.T) {
	nTot := 20
	names := []string{"id", "x", "x2", "v", "id2", "dt", "phi"}
	values := []interface{}{
		[]uint32{ 3, 2, 1 },
		[][3]float32{ {1, 2, 3}, {4, 5, 6}, {7, 8, 9} },
		[][3]float64{ {-1, -1, -1}, {-2, -2, -2}, {-3, -3, -3}},
		[][3]float32{ {3, 3, 3}, {2, 2, 2}, {1, 1, 1} },
		[]uint64{ 10, 20, 30 },
		[]float32{ 1, 1, 1 },
		[]float64{ -10, -100, -1e12 },
	}	
	
	f, err := NewFakeFile(names, values, nTot, binary.BigEndian)
	if err != nil {
		t.Errorf(
			"Expected NewFakeFile() to succeed, but got error '%s'.",
			err.Error(),
		)
	}
	
	hd, err := f.ReadHeader()	
	if err != nil {
		t.Errorf(
			"Expected FakeFile.ReadHeader() to succeed, but got error '%s'.",
			err.Error(),
		)
	} else if err = testFakeFileHeader(hd); err != nil {
		t.Errorf(err.Error())
	} else if err = testFakeFileData(f, hd); err != nil {
		t.Error(err.Error())
	}
}


func testFakeFileHeader(hd Header) error {
	toBytesExp := []byte{4, 8, 15, 16, 23, 42}
	byteOrderExp := binary.BigEndian
	namesExp := []string{"id", "x", "x2", "v", "id2", "dt", "phi"}
	typesExp := []string{ "u32", "v32", "v64", "v32", "u64", "f32", "f64" }
	nTotExp := 20
	zExp := 1.0
	omExp := 0.27
	h100Exp := 0.70
	LExp := 100.0

	if !bytesEq(toBytesExp, hd.ToBytes()) {
		return fmt.Errorf("Expected FakeFileHeader.ToBytes() = %v, got %v.",
			toBytesExp, hd.ToBytes())
	} else if byteOrderExp != hd.ByteOrder() {
		return fmt.Errorf("Expected FakeFileHeader.ByteOrder() = %v, got %v.",
			byteOrderExp, hd.ByteOrder())
	} else if !stringsEq(namesExp, hd.Names()) {
		return fmt.Errorf("Expected FakeFileHeader.Names() = %v, got %v.",
			namesExp, hd.Names())
	} else if !stringsEq(typesExp, hd.Types()) {
		return fmt.Errorf("Expected FakeFileHeader.Types() = %v, got %v.",
			namesExp, hd.Types())
	} else if hd.NTot() != nTotExp {
		return fmt.Errorf("Expected FakeFileHeader.NTot() = %v, got %v.",
			nTotExp, hd.NTot())
	} else if hd.Z() != zExp {
		return fmt.Errorf("Expected FakeFileHeader.Z() = %v, got %v.",
			zExp, hd.Z())
	} else if hd.OmegaM() != omExp {
		return fmt.Errorf("Expected FakeFileHeader.OmegaM() = %v, got %v.",
			omExp, hd.OmegaM())
	} else if hd.H100() != h100Exp {
		return fmt.Errorf("Expected FakeFileHeader.H100() = %v, got %v.",
			zExp, hd.H100())
	} else if hd.L() != LExp {
		return fmt.Errorf("Expected FakeFileHeader.L() = %v, got %v.",
			LExp, hd.L())
	}

	
	return nil
}

func testFakeFileData(f *FakeFile, hd Header) error {
	names := []string{"id", "x", "x2", "v", "id2", "dt", "phi"}
	idExp := []uint32{ 3, 2, 1 }
	xExp := [][3]float32{ {1, 2, 3}, {4, 5, 6}, {7, 8, 9} }
	x2Exp := [][3]float64{ {-1, -1, -1}, {-2, -2, -2}, {-3, -3, -3}}
	vExp := [][3]float32{ {3, 3, 3}, {2, 2, 2}, {1, 1, 1} }
	id2Exp := []uint64{ 10, 20, 30 }
	dtExp := []float32{ 1, 1, 1 }
	phiExp := []float64{ -10, -100, -1e12 }

	buf, err := NewBuffer(hd)
	if err != nil {
		return fmt.Errorf(
			"Expected Buffed could be created, but got error '%s'.",
			err.Error(),
		)
	}
	
	// Look in reverse to make sure the file doesn't assume anything about
	// access order
	for i := range names {
		err := f.Read(names[len(names) - 1 - i], buf)
		if err != nil {
			return fmt.Errorf("Expected '%s' could be read, but go error %s'",
				names[len(names) - 1 - i], err.Error())
		}
	}

	_, _ = xExp, x2Exp
	_, _ = vExp, id2Exp
	_, _, _ = dtExp, phiExp, id2Exp
	if id, _ := buf.Get("id"); !genericEq(id, idExp) {
		return fmt.Errorf("Expected 'id' to be %v, got %v", idExp, id)
	} else if x, _ := buf.Get("x"); !genericEq(x, xExp) {
		return fmt.Errorf("Expected 'x' to be %v, got %v", xExp, x)
	} else if x2, _ := buf.Get("x2"); !genericEq(x2, x2Exp) {
		return fmt.Errorf("Expected 'x2' to be %v, got %v", x2Exp, x2)
	} else if v, _ := buf.Get("v"); !genericEq(v, vExp) {
		return fmt.Errorf("Expected 'v' to be %v, got %v", vExp, v)
	} else if id2, _ := buf.Get("id2"); !genericEq(id2, id2Exp) {
		return fmt.Errorf("Expected 'id2' to be %v, got %v", idExp, id2)
	} else if dt, _ := buf.Get("dt"); !genericEq(dt, dtExp) {
		return fmt.Errorf("Expected 'dt' to be %v, got %v", idExp, dt)
	} else if phi, _ := buf.Get("phi"); !genericEq(phi, phiExp) {
		return fmt.Errorf("Expected 'phi' to be %v, got %v", idExp, phi)
	}

	return nil
}

func genericEq(x, y interface{}) bool {
	switch xx := x.(type) {
	case []float32:
		yy, ok := y.([]float32)
		if !ok { return false }
		return float32sEq(xx, yy) 
	case []float64:
		yy, ok := y.([]float64)
		if !ok { return false }
		return float64sEq(xx, yy) 
	case []uint32:
		yy, ok := y.([]uint32)
		if !ok { return false }
		return uint32sEq(xx, yy) 
	case []uint64:
		yy, ok := y.([]uint64)
		if !ok { return false }
		return uint64sEq(xx, yy) 
	case [][3]float32:
		yy, ok := y.([][3]float32)
		if !ok { return false }
		return vec32sEq(xx, yy) 
	case [][3]float64:
		yy, ok := y.([][3]float64)
		if !ok { return false }
		return vec64sEq(xx, yy)
	default:
		return false
	}
	return false
}

func bytesEq(x, y []byte) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func stringsEq(x, y []string) bool {
		if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
