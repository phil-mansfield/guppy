package snapio

import (
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
					t.Errorf("%d) isRead['%s'] was set to true.".
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
