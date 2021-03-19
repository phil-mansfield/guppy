package snapio

import (
	"bytes"
	"io"
	
	"encoding/binary"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	tests := []struct {
		varNames, varTypes []string
		nf32, nf64, nu32, nu64, nv32, nv64 int
		index []int
		valid bool
	} {
		{[]string{"id"}, []string{"u32"}, 0, 0, 1, 0, 0, 0, []int{0}, true},
		{[]string{"id"}, []string{"u64"}, 0, 0, 0, 1, 0, 0, []int{0}, true},
		{[]string{"x", "id"}, []string{"v32", "u64"}, 0, 0, 0, 1, 1, 0,
			[]int{0, 0}, true},
		{[]string{"x", "id"}, []string{"v64", "u64"}, 0, 0, 0, 1, 0, 1,
			[]int{0, 0}, true},
		{[]string{"x", "id"}, []string{"f32", "u64"}, 1, 0, 0, 1, 0, 0,
			[]int{0, 0}, true},
		{[]string{"x", "id"}, []string{"f64", "u64"}, 0, 1, 0, 1, 0, 0,
			[]int{0, 0}, true},
		{[]string{"x", "id"}, []string{"u32", "u64"}, 0, 0, 1, 1, 0, 0,
			[]int{0, 0}, true},
		{[]string{"x", "id"}, []string{"u64", "u64"}, 0, 0, 0, 2, 0, 0,
			[]int{0, 1}, true},
		{[]string{"x1", "v1", "x2", "phi", "dt", "acc", "id", "id2", "id3"},
			[]string{"v32", "v32", "v64", "f32", "f32", "v32",
				"u32", "u32", "u32"}, 2, 0, 3, 0, 3, 1,
			[]int{0, 1, 0, 0, 1, 2, 0, 1, 2}, true},

		{[]string{}, []string{}, 0, 0, 0, 0, 0, 0, nil, false},
		{[]string{"x"}, []string{"f32"}, 0, 0, 0, 0, 0, 0, nil, false},
		{[]string{"id"}, []string{"f32"}, 0, 0, 0, 0, 0, 0, nil, false},
		{[]string{"id", "x", "x"}, []string{"u32", "f32", "f64"},
			0, 0, 0, 0, 0, 0, nil, false},
		{[]string{"id", "x", "x"}, []string{"u32", "f32", "f34"},
			0, 0, 0, 0, 0, 0, nil, false},
		{[]string{"id", "id"}, []string{"u32", "u32"},
			0, 0, 0, 0, 0, 0, nil, false},
	}

TestLoop:
	for i := range tests {
		buf, err := newBuffer(
			binary.LittleEndian, tests[i].varNames, tests[i].varTypes,
		)

		if err == nil {
			for j := range tests[i].varNames {
				if varType := buf.varType[tests[i].varNames[j]]; varType !=
					tests[i].varTypes[j] {
					t.Errorf("%d) Expected '%s' would have type '%s', got '%s'",
						i, tests[i].varNames[j], tests[i].varTypes[j], varType)
					continue TestLoop
				} else if index := buf.index[tests[i].varNames[j]]; index !=
					tests[i].index[j] {
					t.Errorf("%d) Expected '%s' would have index %d, got %d",
						i, tests[i].varNames[j], tests[i].index[j], index)
					continue TestLoop
				} else if buf.isRead[tests[i].varNames[j]] {
					t.Errorf("%d) isRead['%s'] was set to true.",
						i, tests[i].varNames[j])
				}
			}
		}

		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected varNames = %s, varTypes = %s would succeed, but got error '%s'.",
				i, tests[i].varNames, tests[i].varTypes, err.Error())
		} else if !tests[i].valid && err == nil {
			t.Errorf("%d) Expected varNames = %s, varTypes = %s would fail, but got no error.",
				i, tests[i].varNames, tests[i].varTypes)
		} else if buf == nil {
			continue
		} else if buf.byteOrder != binary.LittleEndian {
			t.Errorf("%d) Expected byteOrder to be %d, but got %d.",
				i, binary.LittleEndian, buf.byteOrder)
		} else if tests[i].nf32 != len(buf.f32)  {
			t.Errorf("%d) Expected len(buf.f32) = %d, but got %d.",
				i, tests[i].nf32, len(buf.f32))
		} else if tests[i].nf64 != len(buf.f64)  {
			t.Errorf("%d) Expected len(buf.f64) = %d, but got %d.",
				i, tests[i].nf64, len(buf.f64))
		} else if tests[i].nu32 != len(buf.u32)  {
			t.Errorf("%d) Expected len(buf.u32) = %d, but got %d.",
				i, tests[i].nu32, len(buf.u32))
		} else if tests[i].nu64 != len(buf.u64)  {
			t.Errorf("%d) Expected len(buf.u64) = %d, but got %d.",
				i, tests[i].nu64, len(buf.u64))
		} else if tests[i].nv32 != len(buf.v32)  {
			t.Errorf("%d) Expected len(buf.v32) = %d, but got %d.",
				i, tests[i].nv32, len(buf.v32))
		} else if tests[i].nv64 != len(buf.v64)  {
			t.Errorf("%d) Expected len(buf.v64) = %d, but got %d.",
				i, tests[i].nv64, len(buf.v64))
		}
	}
}

func TestReadPrimitive(t *testing.T) {
	f32 := []float32{ 1.0, 1.333, 2.0 }
	f32Out := []float32{ 0, 0, 0}
	f64 := []float64{ -1e20, 1.444e14, 6.4 }
	f64Out := []float64{ 0, 0, 0}
	u32 := []uint32{ 4, 8, 15, 16, 23, 42 }
	u32Out := []uint32{ 0, 0, 0, 0, 0, 0}
	u64 := []uint64{ 42, 23, 16, 15, 8, 4 }
	u64Out := []uint64{ 0, 0, 0, 0, 0, 0}
	v32 := [][3]float32{{0.0, 0.1, 0.2}, {0.3, 0.4, 0.5}, {0.6, 0.7, 0.8}}
	v32Out := [][3]float32{ {}, {}, {} }
	v64 := [][3]float64{{0, -0.1, -0.2}, {-0.3, -0.4, -0.5}, {-0.6, -0.7, -0.8}}
	v64Out := [][3]float64{ {}, {}, {} }
	
	orders := []binary.ByteOrder{ binary.LittleEndian, binary.LittleEndian }
	for _, order := range orders {
		buf := &Buffer{ byteOrder: order }
		
		rd := fakeReader(order, f32)
		buf.readPrimitive(rd, f32Out)
		if !float32sEq(f32, f32Out) {
			t.Errorf("Wrote f32 %f with byteOrder = %d, read %f.",
				f32, order, f32Out)
		}

		rd = fakeReader(order, f64)
		buf.readPrimitive(rd, f64Out)
		if !float64sEq(f64, f64Out) {
			t.Errorf("Wrote f64 %f with byteOrder = %d, read %f.",
				f64, order, f64Out)
		}

		rd = fakeReader(order, v32)
		buf.readPrimitive(rd, v32Out)
		if !vec32sEq(v32, v32Out) {
			t.Errorf("Wrote v32 %f with byteOrder = %d, read %f.",
				v32, order, v32Out)
		}

		rd = fakeReader(order, v64)
		buf.readPrimitive(rd, v64Out)
		if !vec64sEq(v64, v64Out) {
			t.Errorf("Wrote v64 %f with byteOrder = %d, read %f.",
				v64, order, v64Out)
		}

		rd = fakeReader(order, u32)
		buf.readPrimitive(rd, u32Out)
		if !uint32sEq(u32, u32Out) {
			t.Errorf("Wrote u32 %d with byteOrder = %d, read %d.",
				u32, order, u32Out)
		}

		rd = fakeReader(order, u64)
		buf.readPrimitive(rd, u64Out)
		if !uint64sEq(u64, u64Out) {
			t.Errorf("Wrote u64 %d with byteOrder = %d, read %d.",
				u64, order, u64Out)
		}
	}
}

func fakeReader(order binary.ByteOrder, x interface{}) io.Reader {
	buf := &bytes.Buffer{ }
	binary.Write(buf, order, x)
	b := buf.Bytes()
	rd := bytes.NewReader(b)
	return rd
	
}

///////////////////////
// Utility functions //
///////////////////////

func float32sEq(x, y []float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func float64sEq(x, y []float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func uint32sEq(x, y []uint32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func uint64sEq(x, y []uint64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func vec32sEq(x, y [][3]float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func vec64sEq(x, y [][3]float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
