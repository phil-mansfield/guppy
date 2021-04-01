package compress

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/phil-mansfield/guppy/lib/particles"
)

type TypeFlag byte
const (
	Uint32Flag TypeFlag = iota
	Uint64Flag
	Float32Flag
	Float64Flag
	numFlags
)

type MethodFlag byte
const (
	LagrangianDeltaFlag MethodFlag = iota
)

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

type Buffer struct {
	b []byte
	u32 []uint32
	u64, u64_2 []uint64
	f32 []float32
	f64 []float64
}

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

	if cap(buf.u64_2) >= n {
		buf.u64_2 = buf.u64_2[:n]
	} else {
		buf.u64_2 = buf.u64_2[:cap(buf.u64_2)]
		buf.u64_2 = append(buf.u64_2, make([]uint64, n - len(buf.u64_2))...)
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

func NewBuffer() *Buffer {
	return &Buffer{ []byte{ }, []uint32{ }, []uint64{ },
		[]uint64{ }, []float32{ }, []float64{ } }
}

func quantize(f particles.Field, delta float64, out []uint64) {
	switch x := f.Data().(type) {
	case []uint64:
		for i := range out { out[i] = x[i] }		
	case []uint32:
		for i := range out { out[i] = uint64(x[i]) }
	case []float32:
		delta32 := float32(0.0)
		for i := range x { out[i] = uint64(x[i] / delta32) }
	case []float64:
		for i := range x { out[i] = uint64(x[i] / delta) }
	default:
		panic("'Impossible' type configuration.")
	}
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
	x := f.Data()
	
	err := binary.Write(wr, m.order, GetTypeFlag(x))
	if err != nil { return err }
	err = binary.Write(wr, m.order, uint32(len(f.Name())))
	if err != nil { return err }
	err = binary.Write(wr, m.order, []byte(f.Name()))
	if err != nil { return err }
	
	err = binary.Write(wr, m.order, x)
	return err
}

func (m *LagrangianDelta) Decompress(
	buf *Buffer, rd io.Reader,
) (particles.Field, error) {
	typeFlag := TypeFlag(0)
	nName := uint32(0)
	
	err := binary.Read(rd, m.order, &typeFlag)
	if err != nil { return nil, err }
	err = binary.Read(rd, m.order, &nName)
	if err != nil { return nil, err }

	bName := make([]byte, nName)
	err = binary.Read(rd, m.order, bName)
	if err != nil { return nil, err }
	name := string(bName)
	
	var x interface{}
	switch typeFlag {
	case Uint32Flag: x = make([]uint32, m.nTot)
	case Uint64Flag: x = make([]uint64, m.nTot)
	case Float32Flag: x = make([]float32, m.nTot)
	case Float64Flag: x = make([]float64, m.nTot)
	default:
		return nil, fmt.Errorf("Type flag for block was %d, but only flags 0 - %d are mapped to types", typeFlag, numFlags)
	}

	return particles.NewGenericField(name, x)
}


