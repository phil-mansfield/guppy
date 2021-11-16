/*package lib contains various functions needed by guppy and mpi_guppy. The
functions in this particular package mainly just utility functions that might
be useful for other programs manually piping output from Guppy. Almost all
of the heavy lifting is done by lib/'s subpackages.
*/
package lib

import (
	"encoding/binary"
	"io"
	
	"reflect"
	"unsafe"
)

var (
	// Version is the version of the software. This can potentially be used
	// to differentiate between breaking changes to the input/output format.
	Version uint64 = 0x1
	RockstarFormatCode uint64 = 0xffffffff00000001
)

// RockstarParticle is a particle with the structure expected by the
// Rockstar halo finder.
type RockstarParticle struct {
	ID uint64 
	X, V [3]float32
}

type PipeHeader struct {
    Version, Format uint64
    N, NTot int64
    Span, Origin, TotalSpan [3]int64
    Z, OmegaM, OmegaL, H100, L, Mass float64
}

func WriteAsBytes(f io.Writer, buf interface{}) error {
	sysOrder := SystemByteOrder()
	switch x := buf.(type) {
	case []uint32: return binary.Write(f, sysOrder, x)
	case []uint64: return binary.Write(f, sysOrder, x)
	case []float32: return binary.Write(f, sysOrder, x)
	case []float64: return binary.Write(f, sysOrder, x)
	case [][3]float32:
		// Go uses the reflect package to write non-primitive data through
		// the binary package. This is slow and makes tons of heap allocations.
		// So you need to be sneaky and "cast" to a primitive array.
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f32x := *(*[]float32)(unsafe.Pointer(&hd))
        err := binary.Write(f, sysOrder, f32x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case [][3]float64:
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f64x := *(*[]float64)(unsafe.Pointer(&hd))
        err := binary.Write(f, sysOrder, f64x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case []RockstarParticle:
		particleSize := int(unsafe.Sizeof(RockstarParticle{ }))
		
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= particleSize
        hd.Cap *= particleSize

		// RockstarParticle fields have inhomogenous sizes, so we need to
		// convert to bytes.
        bx := *(*[]byte)(unsafe.Pointer(&hd))
        _, err := f.Write(bx)

        hd.Len /= particleSize
        hd.Cap /= particleSize

		return err
	}
	
	panic("Internal error: unrecognized type of interal buffer.")
}

func ReadAsBytes(f io.Reader, buf interface{}) error {
	sysOrder := SystemByteOrder()
	switch x := buf.(type) {
	case []uint32: return binary.Read(f, sysOrder, x)
	case []uint64: return binary.Read(f, sysOrder, x)
	case []float32: return binary.Read(f, sysOrder, x)
	case []float64: return binary.Read(f, sysOrder, x)
	case [][3]float32:
		// Go uses the reflect package to write non-primitive data through
		// the binary package. This is slow and makes tons of heap allocations.
		// So you need to be sneaky and "cast" to a primitive array.
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f32x := *(*[]float32)(unsafe.Pointer(&hd))
        err := binary.Read(f, sysOrder, f32x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case [][3]float64:
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f64x := *(*[]float64)(unsafe.Pointer(&hd))
        err := binary.Read(f, sysOrder, f64x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case []RockstarParticle:
		particleSize := int(unsafe.Sizeof(RockstarParticle{ }))
		
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= particleSize
        hd.Cap *= particleSize

		// RockstarParticle fields have inhomogenous sizes, so we need to
		// convert to bytes.
        bx := *(*[]byte)(unsafe.Pointer(&hd))
        _, err := io.ReadFull(f, bx)

        hd.Len /= particleSize
        hd.Cap /= particleSize

		return err
	}
	
	panic("Internal error: unrecognized type of interal buffer.")
}

func SystemByteOrder() binary.ByteOrder {
	// See https://stackoverflow.com/questions/51332658/any-better-way-to-check-endianness-in-go/51332762
	b := [2]byte{ }
	*(*uint16)(unsafe.Pointer(&b[0])) = uint16(0x0001)
	if b[0] == 0 {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}

