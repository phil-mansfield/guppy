package compress

import (
	"encoding/binary"
	"io"
	"math"
	
	"github.com/phil-mansfield/guppy/lib/particles"
)

// TypeFlag is a flag representing an array type.
type TypeFlag byte
const (
	Uint32Flag TypeFlag = iota
	Uint64Flag
	Float32Flag
	Float64Flag
	numFlags
)

// MethodFlag is a flag representing the method used to compress the data.
type MethodFlag uint32
const (
	LagrangianDeltaFlag MethodFlag = iota
)

// GetTypeFlag returns the type flag associated with an array. Only []uint32,
// []uint64, []float32, and []float64 are supported.
func GetTypeFlag(x interface{}) TypeFlag {
	switch x.(type) {
	case []uint32: return Uint32Flag
	case []uint64: return Uint64Flag
	case []float32: return Float32Flag
	case []float64: return Float64Flag
	default:
		panic("'Impossible' type configuration.")
	}
}

// Buffer is an expandable buffer which is used by many of compress's funcitons
// to avoid unneeded heap allocations.
type Buffer struct {
	b []byte
	u32 []uint32
	u64 []uint64
	f32 []float32
	f64 []float64
	q []int64 // This one is specifically for quanitzation.
	rng *RNG
}

// Resize resizes the buffer so its arrays all have length n.
func (buf *Buffer) Resize(n int) {
	// These need to be handled separately because arrays for different types
	// grow at different rates for different types.
	if cap(buf.b) >= n {
		buf.b = buf.b[:n]
	} else {
		buf.b = buf.b[:cap(buf.b)]
		buf.b = append(buf.b, make([]byte, n - len(buf.b))...)
	}
	
	if cap(buf.u32) >= n {
		buf.u32 = buf.u32[:n]
	} else {
		buf.u32 = buf.u32[:cap(buf.u32)]
		buf.u32 = append(buf.u32, make([]uint32, n - len(buf.u32))...)
	}
	
	if cap(buf.u64) >= n {
		buf.u64 = buf.u64[:n]
	} else {
		buf.u64 = buf.u64[:cap(buf.u64)]
		buf.u64 = append(buf.u64, make([]uint64, n - len(buf.u64))...)
	}

	if cap(buf.q) >= n {
		buf.q = buf.q[:n]
	} else {
		buf.q = buf.q[:cap(buf.q)]
		buf.q = append(buf.q, make([]int64, n - len(buf.q))...)
	}

	if cap(buf.f32) >= n {
		buf.f32 = buf.f32[:n]
	} else {
		buf.f32 = buf.f32[:cap(buf.f32)]
		buf.f32 = append(buf.f32, make([]float32, n - len(buf.f32))...)
	}
	
	if cap(buf.f64) >= n {
		buf.f64 = buf.f64[:n]
	} else {
		buf.f64 = buf.f64[:cap(buf.f64)]
		buf.f64 = append(buf.f64, make([]float64, n - len(buf.f64))...)
	}	
}

// NewBuffer creates a new, resizable Buffer,
func NewBuffer(seed uint64) *Buffer {
	return &Buffer{ []byte{ }, []uint32{ }, []uint64{ },
		[]float32{ }, []float64{ }, []int64{ }, NewRNG(seed) }
}

// quantize comverts an array to []uin64 and write it to out. If the array is
// floating point, it is stored to an accuracy of delta.
func quantize(f particles.Field, delta float64, out []int64) {
	// TODO: calling math.Fllor here is much slower than it needs to be.
	// Replace with conditionals.
	switch x := f.Data().(type) {
	case []uint64:
		for i := range out { out[i] = int64(x[i]) }
	case []uint32:
		for i := range out { out[i] = int64(x[i]) }
	case []float32:
		delta32 := float32(delta)
		for i := range x { out[i] = int64(math.Floor(float64(x[i] / delta32))) }
	case []float64:
		for i := range x { out[i] = int64(math.Floor(x[i] / delta)) }
	default:
		panic("'Impossible' type configuration.")
	}
}

// dequantize converts an []int64 array to a different type of array.
// If the output type is floating point, delta*x + delta*uniform(0, 1) is
// used instead. Assumes that buf has been resized to the same length as
// q.
func dequantize(
	name string, q []int64, delta float64, typeFlag TypeFlag, buf *Buffer,
) particles.Field {
	var f particles.Field
	switch typeFlag {
	case Uint32Flag:
		for i := range buf.u32 { buf.u32[i] = uint32(q[i]) }
		f = particles.NewUint32(name, buf.u32)
	case Uint64Flag:
		for i := range buf.u64 { buf.u64[i] = uint64(q[i]) }
		f = particles.NewUint64(name, buf.u64)
	case Float32Flag:
		buf.rng.UniformSequence(buf.f64)
		for i := range buf.f32 {
			buf.f32[i] = float32(delta*(float64(q[i]) + buf.f64[i]))
		}
		f = particles.NewFloat32(name, buf.f32)
	case Float64Flag:
		buf.rng.UniformSequence(buf.f64)
		for i := range buf.f64 {
			buf.f64[i] = delta*(float64(q[i]) + buf.f64[i])
		}
		f = particles.NewFloat64(name, buf.f64)
	default:
		panic("'Impossible' type configuration.")
	}
	
	return f
}

type Method interface {
	// WriteInfo writes initialization information to a Writer.
	WriteInfo(wr io.Writer) error
	// ReadInfo reads initialization information from a Reader.
	ReadInfo(order binary.ByteOrder, rd io.Reader) error

	// Compress compresses the particles in a given field and writes them to
	// a Writer. The buffer buf is used for intermetiate allocations.
	Compress(f particles.Field, buf *Buffer, wr io.Writer) error
	// Decompress decompresses the particles from a Reader and returns a Field
	// containing them. This Field will use the Buffer buf to create the space
	// for the Field, so you need to copy that data elsewhere before calling
	// Decompress again.
	Decompress(buf *Buffer, rd io.Reader) (particles.Field, error)
}

// LagrangianDelta is a compression method which encodes the difference
// between variables along lines in Lagrangian space. It implements the
// Method interface. See the documentation for Method for descriptions of the
// various class methods.
type LagrangianDelta struct {
	order binary.ByteOrder
	span [3]int
	nTot, dir int
	delta float64
}

// NewLagrangianDelta creates a new LagrangianDelta object using a given byte
// ordering. The data inside will have dimensions given by span, encoding
// will be done along the dimension, dir (0 -> x etc.), and floating point
// data will be encoded such that values are stored to within at least delta
// of their starting positions.
func NewLagrangianDelta(
	order binary.ByteOrder, span [3]int, dir int, delta float64,
) *LagrangianDelta {
	nTot := span[0]*span[1]*span[2]
	return &LagrangianDelta{ order, span, nTot, dir, delta }
}

func (m *LagrangianDelta) WriteInfo(wr io.Writer) error {
	span64 := [3]uint64{uint64(m.span[0]), uint64(m.span[1]), uint64(m.span[2])}
	
	err := binary.Write(wr, m.order, span64)
	if err != nil { return err}
	err = binary.Write(wr, m.order, uint64(m.dir))
	if err != nil { return err }
	err = binary.Write(wr, m.order, m.delta)
	return err
}

func (m *LagrangianDelta) ReadInfo(order binary.ByteOrder, rd io.Reader) error {
	m.order = order
	
	span64 := [3]uint64{ }
	dir64 := uint64(0)

	err := binary.Read(rd, m.order, &span64)
	if err != nil { return err }
	err = binary.Read(rd, m.order, &dir64)
	if err != nil { return err }
	err = binary.Read(rd, m.order, &m.delta)
	if err != nil {return err }

	m.span = [3]int{ int(span64[0]), int(span64[1]), int(span64[2]) }
	m.dir = int(dir64)
	m.nTot = m.span[0]*m.span[1]*m.span[2]
	return nil
}

func (m *LagrangianDelta) Compress(
	f particles.Field, buf *Buffer, wr io.Writer,
) error {
	buf.Resize(f.Len())
	x := f.Data()
	
	err := binary.Write(wr, m.order, GetTypeFlag(x))
	if err != nil { return err }
	err = binary.Write(wr, m.order, uint32(len(f.Name())))
	if err != nil { return err }
	err = binary.Write(wr, m.order, []byte(f.Name()))
	if err != nil { return err }
	
	quantize(f, m.delta, buf.q)
	err = binary.Write(wr, m.order, buf.q)
	return err
}

func (m *LagrangianDelta) Decompress(
	buf *Buffer, rd io.Reader,
) (particles.Field, error) {
	buf.Resize(m.nTot)
	
	var (
		typeFlag TypeFlag
		nName uint32
	)
	
	err := binary.Read(rd, m.order, &typeFlag)
	if err != nil { return nil, err }
	err = binary.Read(rd, m.order, &nName)
	if err != nil { return nil, err }

	bName := make([]byte, nName)
	err = binary.Read(rd, m.order, bName)
	if err != nil { return nil, err }
	name := string(bName)

	err = binary.Read(rd, m.order, buf.q)
	if err != nil { return nil, err }
	
	return dequantize(name, buf.q, m.delta, typeFlag, buf), nil
}


