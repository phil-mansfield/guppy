package snapio

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Buffer struct {
	byteOrder binary.ByteOrder
	
	varType map[string]string
	index map[string]int
	
	f32 [][]float32
	f64 [][]float64

	v32 [][][3]float32
	v64 [][][3]float64
	
	u32 [][]uint32
	u64 [][]uint64
	id []int
}

// newBuffer returns a Buffer object that can read a set of variables with the
// specified types ("f32", "f64", "u32", "u64", "v32", "v64" for floats, uints,
// and 3-vectors with 32- and 64-bit widths, respectively). The byte order of
// the files this buffer will be used to read needs to also be specified.
// Variable names cannot be used more than once, and "id" must be specified
// and it must be "u32" or "u64".
func newBuffer(
	byteOrder binary.ByteOrder, varNames, varTypes []string,
) (*Buffer, error) {
	buf := &Buffer{
		byteOrder: byteOrder, varType: map[string]string{ },
		index: map[string]int{ },
	}
	
	for i, name := range varNames {
		if _, ok := buf.varType[name]; ok {
			return nil, fmt.Errorf(
				"The property name '%s' is used more than once.", name,
			)
		} else if name == "id" && varTypes[i] != "u32" && varTypes[i] != "u64" {
			return nil, fmt.Errorf(
				"'id' is associated with '%s', which is not an integer type.",
				varTypes[i],
			)
		}
		
		buf.varType[name] = varTypes[i]
		
		switch varTypes[i] {
		case "f32":
			buf.f32 = append(buf.f32, []float32{ })
			buf.index[name] = len(buf.f32) - 1
		case "f64":
			buf.f64 = append(buf.f64, []float64{ })
			buf.index[name] = len(buf.f64) - 1
		case "v32":
			buf.v32 = append(buf.v32, [][3]float32{ })
			buf.index[name] = len(buf.v32) - 1
		case "v64":
			buf.v64 = append(buf.v64, [][3]float64{ })
			buf.index[name] = len(buf.v64) - 1
		case "u32":
			buf.u32 = append(buf.u32, []uint32{ })
			buf.index[name] = len(buf.u32) - 1
		case "u64":
			buf.u64 = append(buf.u64, []uint64{ })
			buf.index[name] = len(buf.u64) - 1
		default:
			return nil, fmt.Errorf("'%s' is not a valid type. Only 'f32', 'f64', 'v32', 'v64', 'u32', and 'u64' are valid.", varTypes[i])
		}
	}

	if _, ok := buf.varType["id"]; !ok {
		return nil, fmt.Errorf("No 'id' property was specified.")
	}
	
	return buf, nil
}

// Reset resets a buffer so that a new file can be read into it. This allows
// informative internal errors to be thrown.
func (buf *Buffer) Reset() {
}

// read reads the data associated with a given variable name to Buffer. n values
// are read.
func (buf *Buffer) read(rd io.Reader, name string, n int) error {
	panic("NYI")
}

// Get returns an interface pointing to the slice associated with a given
// variable name.
func (buf *Buffer) Get(rd io.Reader, name string) (interface{}, error) {
	panic("NYI")
}
