package snapio

/* This file handles codes related to the generic Buffer object. Much of this
code is just type switches. It'll be obsolete if Guppy is ever ported to Go 2.

After writing and testing this, I realized that the code can be made much
simpler if Buffer just has an array of interface{} values instead of different
arrays for each type, so if someone does a rewrite of this, just do that.
*/

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

type Buffer struct {
	byteOrder binary.ByteOrder
	
	varType map[string]string
	index map[string]int
	isRead map[string]bool
	
	f32 [][]float32
	f64 [][]float64

	v32 [][][3]float32
	v64 [][][3]float64
	
	u32 [][]uint32
	u64 [][]uint64
	id []int

	
}

// NewBuffer creates a buffer object which can read files with the given Header.
func NewBuffer(hd Header) (*Buffer, error) {
	return newBuffer(hd.ByteOrder(), hd.Names(), hd.Types())
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
		index: map[string]int{ }, isRead: map[string]bool{ },
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
		buf.isRead[name] = false
		
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
	for name := range buf.isRead {
		buf.isRead[name] = false
	}
}

// read reads the data associated with a given variable name to Buffer. n values
// are read.
func (buf *Buffer) read(rd io.Reader, name string, n int) error {
	varType, ok := buf.varType[name]
	if !ok {
		return fmt.Errorf("The property name '%s' hasn't been registered to the file.", name)
	}

	if buf.isRead[name] {
		return fmt.Errorf("The property name '%s' is being read multiple times without a call to Reset().", name)
	}

	i := buf.index[name]
	var err error
	switch varType {
	case "f32":
		buf.f32[i], _ = expand(buf.f32[i], n).([]float32)
		err = buf.readPrimitive(rd, buf.f32[i])
	case "f64":
		buf.f64[i], _ = expand(buf.f64[i], n).([]float64)
		err = buf.readPrimitive(rd, buf.f64[i])
	case "u32":
		buf.u32[i], _ = expand(buf.u32[i], n).([]uint32)
		err = buf.readPrimitive(rd, buf.u32[i])
	case "u64":
		buf.u64[i], _ = expand(buf.u64[i], n).([]uint64)
		err = buf.readPrimitive(rd, buf.u64[i])
	case "v32":
		buf.v32[i], _ = expand(buf.v32[i], n).([][3]float32)
		err = buf.readPrimitive(rd, buf.v32[i])
	case "v64":
		buf.v64[i], _ = expand(buf.v64[i], n).([][3]float64)
		err = buf.readPrimitive(rd, buf.v64[i])
	default:
		return fmt.Errorf("'%s' is not a valid type. Only 'f32', 'f64', 'v32', 'v64', 'u32', and 'u64' are valid.", varType)
	}

	buf.isRead[name] = true
	
	return err
}

// readPrimitive reads data from a reader into x, an interface around an array.
// Supported types are []float32, []float64, []uint32, []uint64, [][3]float32,
// [][3]float64. Returns an error if given an unsupported type or an I/O error.
func (buf *Buffer) readPrimitive(rd io.Reader, x interface{}) error {
	var err error
	switch xx := x.(type) {
	case []float32: err = binary.Read(rd, buf.byteOrder, xx)
	case []float64: err = binary.Read(rd, buf.byteOrder, xx)
	case []uint32: err = binary.Read(rd, buf.byteOrder, xx)
	case []uint64: err = binary.Read(rd, buf.byteOrder, xx)
	case [][3]float32:
		// This is done this way because binary.Read does a bunch of heap
		// allocations when used on [][3]float32 arrays.		
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&xx))
		hd.Len *= 3
		hd.Cap *= 3
		
		f32x := *(*[]float32)(unsafe.Pointer(&hd))
		err = binary.Read(rd, buf.byteOrder, f32x)
		
		hd.Len /= 3
		hd.Cap /= 3
	case [][3]float64:
		// This is done this way because binary.Read does a bunch of heap
		// allocations when used on [][3]float32 arrays.
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&xx))
		hd.Len *= 3
		hd.Cap *= 3
		
		f64x := *(*[]float64)(unsafe.Pointer(&hd))
		err = binary.Read(rd, buf.byteOrder, f64x)
		
		hd.Len /= 3
		hd.Cap /= 3
	default:
		panic("(Supposedly) impossible type configuration")
	}
	return err
}

// expand expands an array to have size n.
func expand(x interface{}, n int) interface{} {
	switch xx := x.(type) {
	case []float32:
		m := len(xx)
		if m < n { xx = append(xx, make([]float32, n-m)...) }
		return xx[:n]
	case []float64:
		m := len(xx)
		if m < n { xx = append(xx, make([]float64, n-m)...) }
		return xx[:n]
	case [][3]float32:
		m := len(xx)
		if m < n { xx = append(xx, make([][3]float32, n-m)...) }
		return xx[:n]
	case [][3]float64:
		m := len(xx)
		if m < n { xx = append(xx, make([][3]float64, n-m)...) }
		return xx[:n]
	case []uint32:
		m := len(xx)
		if m < n { xx = append(xx, make([]uint32, n-m)...) }
		return xx[:n]
	case []uint64:
		m := len(xx)
		if m < n { xx = append(xx, make([]uint64, n-m)...) }
		return xx[:n]
	}
	panic("(Supposedly) impossible type configuration.")
}

// Get returns an interface pointing to the slice associated with a given
// variable name.
func (buf *Buffer) Get(name string) (interface{}, error) {
	varType, ok := buf.varType[name]
	if !ok {
		return nil, fmt.Errorf("'%s' is not a recognized variable name.", name)
	} else if !buf.isRead[name] {
		return nil, fmt.Errorf("'%s' has not been read.", name)
	}

	idx := buf.index[name]
	switch varType {
	case "f32": return buf.f32[idx], nil
	case "f64": return buf.f64[idx], nil
	case "v32": return buf.v32[idx], nil
	case "v64": return buf.v64[idx], nil
	case "u32": return buf.u32[idx], nil
	case "u64": return buf.u64[idx], nil
	}

	panic("(Supposedly) impossible type configuration")
}
