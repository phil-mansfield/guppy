package compress

import (
	"encoding/binary"
	"io"
	"math"
	"compress/zlib"
	"fmt"
	"strings"
	"bytes"

	"github.com/phil-mansfield/guppy/lib/particles"
	"github.com/DataDog/zstd"
)

// TypeFlag is a flag representing an array type.
type TypeFlag int64
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

// Buffer is an expandable buffer which is used by many of compress's functions
// to avoid unneeded heap allocations.
type Buffer struct {
	b []byte
	u32 []uint32
	u64 []uint64
	f32 []float32
	f64 []float64

	// These two buffers are specifically for quantiazation and encoding.
	bZStd []byte
	i64 []int64
	q []int64
	rng *RNG
}

// Resize resizes the buffer so its arrays all have length n.
func (buf *Buffer) Resize(n int) {
	// These need to be handled separately because arrays for different types
	// grow at different rates.
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

	if cap(buf.i64) >= n {
		buf.i64 = buf.i64[:n]
	} else {
		buf.i64 = buf.i64[:cap(buf.i64)]
		buf.i64 = append(buf.i64, make([]int64, n - len(buf.i64))...)
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
		[]float32{ }, []float64{ }, []byte{ }, []int64{ }, []int64{ },
		NewRNG(seed) }
}

// quantize comverts an array to []uin64 and write it to out. If the array is
// floating point, it is stored to an accuracy of delta.
func Quantize(f particles.Field, delta float64, qPeriod int64, out []int64) {
	// TODO: calling math.Fllor here is much slower than it needs to be.
	// Replace with conditionals.
	switch x := f.Data().(type) {
	case []uint64:
		for i := range out { out[i] = int64(x[i]) }
	case []uint32:
		for i := range out { out[i] = int64(x[i]) }
	case []float32:
		delta32 := float32(delta)
		for i := range x {
			out[i] = int64(math.Floor(float64(x[i] / delta32)))
		}
	case []float64:
		for i := range x {
			out[i] = int64(math.Floor(x[i] / delta))
		}
	default:
		panic("'Impossible' type configuration.")
	}

	if qPeriod > 0 {
		for i := range out {
			if out[i] < 0 {
				out[i] += qPeriod
			} else if out[i] >= qPeriod {
				out[i] -= qPeriod
			}
		}
	}
}

// dequantize converts an []int64 array to a different type of array.
// If the output type is floating point, delta*x + delta*uniform(0, 1) is
// used instead. Assumes that buf has been resized to the same length as
// q.
func Dequantize(
	name string, q []int64, delta float64, qPeriod int64,
	typeFlag TypeFlag, buf *Buffer,
) particles.Field {

	if qPeriod > 0 {
		for i := range q {
			// This can be slow on some systems, where for jumps backwards
			// always predict true. It's neccessary because you have no
			// garuantee that the deltas add up to a value inside the box.
			// If this ends up being slow, you'll need to refactor and do this
			// check in-line with the delta calculation.
			for q[i] < 0 {
				q[i] += qPeriod
			}
			for q[i] >= qPeriod {
				q[i] -= qPeriod
			}
		}
	}

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

// Method is an interface representing a compression method.
type Method interface {
	// MethodFlag returns the method used to compress the data.
	MethodFlag() MethodFlag
	// SetOrder sets the byte order of the compression method.
	SetOrder(order binary.ByteOrder)
	// Span returns the span of the data compressed by the method.
	Span() [3]int

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
	Decompress(buf *Buffer, rd io.Reader, name string) (particles.Field, error)
}

// LagrangianDelta is a compression method which encodes the difference
// between variables along lines in Lagrangian space. It implements the
// Method interface. See the documentation for Method for descriptions of the
// various class methods.
type LagrangianDelta struct {
	order binary.ByteOrder
	span [3]int
	nTot int
	delta, period float64
}

// NewLagrangianDelta creates a new LagrangianDelta object. The span of the
// particles in ID-space is given by span, the minimum accuracy is given by
// delta, and the periodicity is given by period (i.e. "the size of the box").
// If this method is being used on non periodic data set period to a
// non-positive number.
func NewLagrangianDelta(span [3]int, delta, period float64) *LagrangianDelta {
	nTot := span[0]*span[1]*span[2]
	if period > 0 {
		nPix := math.Ceil(period / delta)
		delta = period / nPix
	}
	return &LagrangianDelta{ binary.LittleEndian, span, nTot, delta, period }
}

// (see documentaion for the Method interface)
func (m *LagrangianDelta) SetOrder(order binary.ByteOrder) { m.order = order }

// (see documentaion for the Method interface)
func (m *LagrangianDelta) MethodFlag() MethodFlag {
	return LagrangianDeltaFlag
}

// (see documentaion for the Method interface)
func (m *LagrangianDelta) Span() [3]int { return m.span }

// (see documentaion for the Method interface)
func (m *LagrangianDelta) WriteInfo(wr io.Writer) error {
	span64 := [3]uint64{uint64(m.span[0]), uint64(m.span[1]), uint64(m.span[2])}
	
	err := binary.Write(wr, m.order, LagrangianDeltaFlag)
	if err != nil { return err }
	err = binary.Write(wr, m.order, span64)
	if err != nil { return err }
	err = binary.Write(wr, m.order, m.delta)
	return err
}

// (see documentaion for the Method interface)
func (m *LagrangianDelta) ReadInfo(order binary.ByteOrder, rd io.Reader) error {
	var flag MethodFlag
	err := binary.Read(rd, order, &flag)
	if flag != LagrangianDeltaFlag {
		return fmt.Errorf("Mismatch between the Method type used to " + 
			"decompress block and the Method type used to compress it. Block " +
			"was compressed with LagrangianDelta (flag = %d), but block flag " +
			"was %d.", LagrangianDeltaFlag, flag)
	}

	m.order = order
	span64 := [3]uint64{ }

	err = binary.Read(rd, m.order, &span64)
	if err != nil { return err }
	err = binary.Read(rd, m.order, &m.delta)
	if err != nil {return err }

	m.span = [3]int{ int(span64[0]), int(span64[1]), int(span64[2]) }
	m.nTot = m.span[0]*m.span[1]*m.span[2]
	return nil
}

// (see documentaion for the Method interface)
func (m *LagrangianDelta) Compress(
	f particles.Field, buf *Buffer, wr io.Writer,
) error {
	buf.Resize(f.Len())

	typeFlag := GetTypeFlag(f.Data())

	qPeriod := int64(0)
	if m.period > 0 || m.delta > 0 {
		qPeriod = int64(math.Ceil(m.period/m.delta))
	}
	
	Quantize(f, m.delta, qPeriod, buf.q)
	
	firstDim := ChooseFirstDim(f.Name())
	slices := BlockToSlices(m.span, firstDim, buf.q, buf.i64)
	offsets := SliceOffsets(slices)
	
	// Replace each slice with its deltas. Remember, this modifies buf.i64.
	for i := range slices {
		DeltaEncode(offsets[i], qPeriod, slices[i], slices[i])
	}
	
	stats := &DeltaStats{ }
	stats.Load(buf.i64)
	mid := stats.Window(256)
	rot := stats.NeededRotation(mid)
	
	RotateEncode(buf.i64, rot)

	hd := &lagrangianDeltaHeader{ typeFlag, buf.q[0], rot }
	err := binary.Write(wr, m.order, hd)
	if err != nil { return err }

	// Write to disk.
	buf.bZStd, err = WriteCompressedIntsZStd(buf.i64, buf.b, buf.bZStd, wr)
	if err != nil {
		return fmt.Errorf("zlib error while writing block '%s': %s",
			f.Name(), err.Error())
	}

	return err
}

// CooseFirstDim chooses the first encoded dimension for a variable with a
// given name. This is chosen so almost all the deltas are perpendicular to the
// direction of the vector if the stored data is vector.
func ChooseFirstDim(name string) int {
	switch {
	case strings.Index(name, "[0]") >= 0: return 0
	case strings.Index(name, "[1]") >= 0: return 1
	case strings.Index(name, "[2]") >= 0: return 2
	default: return 0
	}
}

// lagrangianDeltaHeader is the header that LagrangianDelta writes to disk
// before writing the data block.
type lagrangianDeltaHeader struct {
	TypeFlag TypeFlag
	FirstOffset, Rot int64
}

// (see documentaion for the Method interface)
func (m *LagrangianDelta) Decompress(
	buf *Buffer, rd io.Reader, name string,
) (particles.Field, error) {	
	buf.Resize(m.nTot)

	hd := &lagrangianDeltaHeader{ }
	err := binary.Read(rd, m.order, hd)
	if err != nil { return  nil, err}

	firstDim := ChooseFirstDim(name)

	// Read data. This is done by adding bytes to buf.i64 one-by-one, so
	// we need to clear the array first.
	for i := range buf.i64 { buf.i64[i] = 0 }
	buf.b, buf.bZStd, err = ReadCompressedIntsZStd(
		rd, buf.b, buf.bZStd, buf.i64)

	
	if err != nil {
		return nil, fmt.Errorf("zlib error while reading block '%s': %s",
			name, err.Error())
	}

	qPeriod := int64(0)
	if m.period > 0 && m.delta > 0 {
		qPeriod = int64(math.Ceil(m.period/m.delta))
	}

	// Invert the procedures used in Compress.
	slices := MakeDeltaSlices(m.span, firstDim, buf.i64)
	RotateDecode(buf.i64, hd.Rot)
	DeltaDecodeFromSlices(hd.FirstOffset, slices)
	SlicesToBlock(m.span, firstDim, slices, buf.q)
	return Dequantize(name, buf.q, m.delta, qPeriod, hd.TypeFlag, buf), nil
}

// intToByte transfers a one-byte "column" from u64 to b. The bytes are indexed
// from least to most significant. int64's are used here, but they will always
// be positive.
func intToByte(i64 []int64, b []byte, col int) {
	for i := range i64 {
		b[i] = byte((uint64(i64[i]) >> (8*col)) & 0xff)
	}
}

// byteToInt adds a one-byte column
func byteToInt(b []byte, i64 []int64, col int) {
	for i := range i64 {
		i64[i] += int64(uint64(b[i]) << (8*col))
	}
}

// reszieBytes resizes a byte buffer to have length n.
func resizeBytes(b []byte, n int) []byte {
	if cap(b) >= n {
		b = b[:n]
	} else {
		b = b[:cap(b)]
		b = append(b, make([]byte, n - len(b))...)
	}

	return b
}

// writeCompressedIntsZlib writes an array of ints, q, to an io.Writer using
// column-ordered zlib blocks. b is used as a temporary internal buffer 
// and must be the same length as q.
//
// This function is based on zlib entropy encoding
func WriteCompressedIntsZLib(q []int64, b []byte, wr io.Writer) error {
	if len(q) != len(b) {
		panic(fmt.Sprintf("Internal error: output byte buffer has length %d," + 
			" but quantized int array had length %d.", len(b), len(q)))
	}
	
	for i := 0; i < 8; i++ {
		// We need to create a new Writer each loop so a different codex is
		// used for the different columns, letting the high-significance bits
		// be compressed to basically nothing.
		wrZLib := zlib.NewWriter(wr)

		intToByte(q, b, i)
		_, err := wrZLib.Write(b)
		if err != nil { return err }

		err = wrZLib.Close()
		if err != nil { return err }
	}
	
	return nil
}

// readCompressedIntsZLib reads an array of ints, q, from an io.Reader using
// column-ordered zlib blocks. b is used as a temporary internal buffer and
// will be resized as needed. A resized version is returned by the
// function.
//
// This function is based on zlib entropy encoding
func ReadCompressedIntsZLib(rd io.Reader, b []byte, q []int64) ([]byte, error) {
	buf := bytes.NewBuffer(b[:0])	
	for i := 0; i < 8; i++ {
		buf.Reset()

		rdZLib, err := zlib.NewReader(rd)
		if err != nil { return nil, err }

		_, err = io.Copy(buf, rdZLib)
		if err != nil { return nil, err }

		b = buf.Bytes()
		byteToInt(b, q, i)

		err = rdZLib.Close()
		if err != nil { return nil, err }
	}

	return b, nil
}

// writeCompressedIntsZStd writes an array of ints, q, to an io.Writer using
// column-ordered zlib blocks. b is used as a temporary internal buffer 
// and must be the same length as q. buf is a buffer used internally and will
// be resized as needed and returned. Just keep passing the same buffer to
// the WriteCompressedIntsZLib function and you'll be okay.
//
// This function is based on zstd entropy encoding
func WriteCompressedIntsZStd(
	q []int64, b, buf []byte, wr io.Writer,
) ([]byte, error) {
	
	if len(q) != len(b) {
		panic(fmt.Sprintf("Internal error: output byte buffer has length %d,"+ 
			" but quantized int array had length %d.", len(b), len(q)))
	}

	for i := 0; i < 8; i++ {
		// We need to create a new Writer each loop so a different codex is
		// used for the different columns, letting the high-significance bits
		// be compressed to basically nothing.
		intToByte(q, b, i)

		var err error
		buf, err = zstd.CompressLevel(buf, b, 1)
		if err != nil { return nil, err }

		err = binary.Write(wr, binary.LittleEndian, int64(len(buf)))
		if err != nil { return nil, err }
		
		_, err = wr.Write(buf)
		if err != nil { return nil, err }
	}
	
	// Resize so I don't accidentally do something dumb with the buffer.
	return buf[:0], nil
}

// readCompressedIntsZLib reads an array of ints, q, from an io.Reader using
// column-ordered zlib blocks. b and buf are used as a temporary internals
// buffers and  will be resized as needed. Resized versions are returned by the
// function.
//
// This function is based on zstd entropy encoding.
func ReadCompressedIntsZStd(
	rd io.Reader, b, buf []byte, q []int64,
) (bOut, bufOut []byte, err error) {
	b = resizeBytes(b, len(q))
	
	for i := 0; i < 8; i++ {
		nBuf := int64(0)
		err := binary.Read(rd, binary.LittleEndian, &nBuf)
		if err != nil { return nil, nil, err }
		
		buf = resizeBytes(buf, int(nBuf))
		_, err = io.ReadFull(rd, buf)
		
		b, err = zstd.Decompress(b, buf)
		if err != nil { return nil, nil, err }
		
		byteToInt(b, q, i)
	}

	// Resize so I don't accidentally do something dumb with the buffers.
	return b[:0], buf[:0], nil
}

// splitArray splits the array x into n smaller arrays and writes their slice
// headers to splits. The length of each array is given by lengths. n =
// len(lengths) = len(splits).
func splitArray(x []int64, lengths []int, splits [][]int64) {
	sum := 0
	for _, n := range lengths { sum += n }

	if sum != len(x) {
		panic(fmt.Sprintf("Internal error: sum of length = %d, but length " + 
			"of array is %d.", sum, len(x)))
	} else if len(lengths) != len(splits) {
		panic(fmt.Sprintf("Internal error: len(lengths) = %d, len(splits) " + 
			"= %d.", len(lengths), len(splits)))
	}

	start := 0
	for i := range lengths {
		end := start + lengths[i]
		splits[i] = x[start: end]
		start = end
	}
}

// DeltaEncode delta encodes the array x into the array out. The element
// before x[0] is taken to be offset. x and out can be the same array. Since the
// encoding is done into a uint64 array,
func DeltaEncode(offset, qPeriod int64, x, out []int64) {
	if len(x) != len(out) {
		panic(fmt.Sprintf("Internal error: len(x) = %d, but len(out) = " + 
			"%d in deltaEncode", len(x), len(out)))
	}
	if len(x) == 0 { return }

	// This is a bit wonky, but you need to do the loop this way to allow
	// deltaEncode to be called in place.

	prev := x[0]
	out[0] = prev - offset
	for i := 1; i < len(x); i++ {
		next := x[i]
		out[i] = next - prev
		prev = next
	}

	if qPeriod > 0 {
		for i := range out {
			if out[i] > qPeriod/2 {
				out[i]  -= qPeriod
			} else if out[i] < -qPeriod/2 {
				out[i] += qPeriod
			}
		}
	}
}

// DeltaDecode decodes a integer array encoded with DeltaEncode.
func DeltaDecode(offset int64, x, out []int64) {
	if len(x) != len(out) {
		panic(fmt.Sprintf("Internal error: len(x) = %d, but len(out) = " + 
			"%d in deltaDecode", len(x), len(out)))
	}
	if len(x) == 0 { return }

	out[0] = offset + x[0]
	for i := 1; i < len(out); i++ {
		out[i] = out[i-1] + x[i]
	}
}

// BlockToSlices converts a block of x-major indices into a set of slices
// which each correspond to a 1-dimensional "skewer" through the block. These
// are organized so only one actual value needs to be stored for the block.
// First, one skewer down firstDim, then a face of skewers in the next
// direciton, and then a block of skewers filling out the rest of the data.
func BlockToSlices(span [3]int, firstDim int, x, buf []int64) [][]int64 {
	if len(buf) != len(x) {
		panic(fmt.Sprintf("Internal error: len(x) = %d, but len(buf) = %d",
			len(x), len(buf)))
	}

	// You'll probably want a pen and paper when trying to understand this
	// function. I did my best, sorry!

	out := MakeDeltaSlices(span, firstDim, buf)

	dx := [3]int{ 1, span[0], span[0]*span[1] }

	d1, d2, d3 := firstDim, (firstDim + 1) % 3, (firstDim + 2) % 3

	// i* indexes over the first index of the skewer
	// j indexes along the skewer
	// k indexes along the output slices

	// Copy over the first skewer.
	start := 0*dx[d1] + 0*dx[d2] + 0*dx[d3]
	for j := 0; j < span[d1]; j++ {
		out[0][j] = x[start + dx[d1]*j] 
	}

	// Copy over the first face.
	for i1 := 0; i1 < span[d1]; i1++ {
		k := i1 + 1

		start := i1*dx[d1] + 0*dx[d2] + 0*dx[d3]
		for i2 := 1; i2 < span[d2]; i2++ {
			j := i2 - 1 
			out[k][j] = x[start + dx[d2]*i2]
		}
	}

	// Copy over the body of the block.
	for i2 := 0; i2 < span[d2]; i2++ {
		for i1 := 0; i1 < span[d1]; i1++ {

			start := i1*dx[d1] + i2*dx[d2] + 1*dx[d3]
			k := 1 + span[d1] + i1 + i2*span[d1]

			for i3 := 1; i3 < span[d3]; i3++ {
				j := i3 - 1
				out[k][j] = x[start + dx[d3]*j]
			}
		}
	}

	return out
}

// MakeDeltaSlices splits up an array, buf, into slices according to the
// splitting strategy used by BlockToSlices: first slice has length
// span[firstDim], next span[firstDim] sliaces have length span[secondDim] - 1,
// next span[firstDim]*span[secondDim] have length span[thridDim] - 1. 
func MakeDeltaSlices(span [3]int, firstDim int, buf []int64) [][]int64 {
	lens := sliceLengths(span, firstDim)
	out := make([][]int64, len(lens))
	splitArray(buf, lens, out)
	return out
}

// SliceOffsets returns the offset associated with each slice within the
// overall block.
func SliceOffsets(x [][]int64) []int64 {
	offsets := make([]int64, len(x))

	offsets[0] = x[0][0]
	for i := range x[0] {
		offsets[i + 1] = x[0][i]
		offsets[i + len(x[0]) + 1] = x[0][i]
	}

	for j := range x[1] {
		for i := range x[0] {
			offsets[1 + (i + (j+2)*len(x[0]))] = x[1 + i][j]
		}
	}

	return offsets
}

// DeltaDecodeFromSlices runs DeltaDecode on a set of slices. This includes
// finding the correct offsets.
func DeltaDecodeFromSlices(firstOffset int64, x [][]int64) {
	DeltaDecode(firstOffset, x[0], x[0])

	n := len(x[0])
	for i := range x[0] {
		DeltaDecode(x[0][i], x[i + 1], x[i + 1])
		DeltaDecode(x[0][i], x[i + n + 1], x[i + n + 1])
	}

	for j := range x[1] {
		for i := range x[0] {
			slice := x[1 + i + (j+2)*n]
			DeltaDecode(x[1 + i][j], slice, slice)
		}
	}
}

// nSlices returns the number of slices used by LagrangianDelta for a given
// span and starting dimension.
func nSlices(span [3]int, firstDim int) int {
	secondDim := (firstDim + 1) % 3
	return 1 + span[firstDim] + span[secondDim]*span[firstDim]
}

// SliceLengths gives the lengths of the slices that a given block would
// be broken into, using firstDim first.
func sliceLengths(span [3]int, firstDim int) []int {
	secondDim, thirdDim := (firstDim + 1) % 3, (firstDim + 2) % 3
	nTot := nSlices(span, firstDim)
	lens := make([]int, nTot)

	// First array goes down the first dimension of the block.
	lens[0] = span[firstDim]

	// Next block fills out the face made by the 1st/2nd dims.
	for i := 1; i < 1 + span[firstDim]; i++ {
		lens[i] = span[secondDim] - 1
	}

	// Lastly, fill out the body of the block.
	for i := 1 + span[firstDim]; i < nTot; i++ {
		lens[i] = span[thirdDim] - 1
	}

	return lens
}

// SlicesToBlock joins a set of slices, x, into a block in out.
func SlicesToBlock(span [3]int, firstDim int, x [][]int64, out []int64) {
	sum := 0
	for i := range x { sum += len(x[i]) }

	if len(out) != sum {
		panic(fmt.Sprintf("Internal error: sum(len(x)) = %d, but len(out) = %d",
			sum, len(out)))
	}

	dx := [3]int{ 1, span[0], span[0]*span[1] }

	d1, d2, d3 := firstDim, (firstDim + 1) % 3, (firstDim + 2) % 3

	// i* indexes over the first index of the skewer
	// j indexes along the skewer
	// k indexes along the output slices

	// Copy over the first skewer.
	start := 0*dx[d1] + 0*dx[d2] + 0*dx[d3]
	for j := 0; j < span[d1]; j++ {
		out[start + dx[d1]*j] = x[0][j]
	}

	// Copy over the first face.
	for i1 := 0; i1 < span[d1]; i1++ {
		k := i1 + 1

		start := i1*dx[d1] + 0*dx[d2] + 0*dx[d3]
		for i2 := 1; i2 < span[d2]; i2++ {
			j := i2 - 1 
			out[start + dx[d2]*i2] = x[k][j]
		}
	}

	// Copy over the body of the block.
	for i2 := 0; i2 < span[d2]; i2++ {
		for i1 := 0; i1 < span[d1]; i1++ {

			start := i1*dx[d1] + i2*dx[d2] + 1*dx[d3]
			k := 1 + span[d1] + i1 + i2*span[d1]

			for i3 := 1; i3 < span[d3]; i3++ {
				j := i3 - 1
				out[start + dx[d3]*j] = x[k][j]
			}
		}
	}
}
