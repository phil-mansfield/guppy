package compress

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"io"

	"github.com/phil-mansfield/guppy/lib/particles"
	"github.com/phil-mansfield/guppy/lib/snapio"

	"unsafe"
)

const (
	// MagicNumber is an arbirary number at the start of all guppy files
	// which should help identify when the code is run on somehting else by
	// accident.
	MagicNumber = 0xbadf00d0
	// ReverseMagicNumber is the magic number if read on a machine with 
	// flipped endianness.
	ReverseMagicNumber = 0xd000fdba
	Version = 1
)

// Writer is a class which handles writing to disk. The pattern is that you
// create a single writer wiht NewWriter, add fields to it with AddField, and
// to finally call Flush() when you want to flush all the buffers and write to
// disk.
type Writer struct {
	Header
	fname string
	buf *Buffer
	order binary.ByteOrder
	methodFlags []uint32
	headerEdges, dataEdges []int64
	header, data *bytes.Buffer
}

// NewWriter creates a Writer targeting a given file and using a given byte
// ordering. Two buffers need to be passed as arguments, a compress.Buffer
// to handle all the internal arrays needed by the compression methods, and
// a byte array that's used to store an in-RAM version of the file. If you
// don't want to make excess heap allocaitons, pass the same array returned by
// Flush(). You can pass the same compress.Buffer each time.
func NewWriter(
	fname string, snapioHeader snapio.Header,
	idOffset uint64, span [3]int64,
	buf *Buffer, b []byte, order binary.ByteOrder,
) *Writer {
	header := bytes.NewBuffer([]byte{ })
	data := bytes.NewBuffer(b[:0]) 
	hd := convertSnapioHeader(snapioHeader, span, idOffset)

	return &Writer{
		*hd, fname, buf, order, []uint32{},
		[]int64{0}, []int64{0},
		header, data,
	}
}


// Add field adds a new field to the file which will be compressed with
// a given method.
func (wr *Writer) AddField(field particles.Field, method Method) error {
	if int64(field.Len()) != wr.N {
		return fmt.Errorf("File stores %d particles, but was given a new " + 
			"field, %s, with %d particles.", wr.N, field.Name(), field.Len())
	}

	method.SetOrder(wr.order)

	err := method.WriteInfo(wr.header)
	if err != nil { return err }

	err = method.Compress(field, wr.buf, wr.data)
	if err != nil { return err }

	wr.headerEdges = append(wr.headerEdges, int64(wr.header.Len()))
	wr.dataEdges = append(wr.dataEdges, int64(wr.data.Len()))
	wr.methodFlags = append(wr.methodFlags, uint32(method.MethodFlag()))

	wr.Names = append(wr.Names, field.Name())
	switch field.Data().(type) {
	case []uint32: wr.Types = append(wr.Types, "u32")
	case []uint64: wr.Types = append(wr.Types, "u64")
	case []float32: wr.Types = append(wr.Types, "f32")
	case []float64: wr.Types = append(wr.Types, "f64")
	default:
		panic(fmt.Sprintf("Internal error: unknown-typed " + 
			"paritcles.Field (name: '%s') given " + 
			"to Writer.AddField()", field.Name()))
	}

	return nil
}

// Flush flushes the internal buffers to disk. It returns a (potentially
// cap-expanded) byte array that can be passed to later call to NewWriter().
func (wr *Writer) Flush() ([]byte, error) {
	fp, err := os.Create(wr.fname)
	if err != nil { return nil, err }
	defer fp.Close()

	// Number of bytes used by the file's header (i.e. not the method headers)
	nHd := 0

	// Write file identificaiton information.
	magicNumber := uint32(MagicNumber)
	version := uint32(Version)

	err = binary.Write(fp, wr.order, magicNumber)
	if err != nil { return nil, err }
	nHd += 4
	err = binary.Write(fp, wr.order, version)
	if err != nil { return nil, err }
	nHd += 4

	n, err := wr.Header.write(fp, wr.order)
	if err != nil { return nil, err }
	nHd += n

	// Write the actual navigation information.
	nHd += 4*len(wr.methodFlags) // methodFalgs size
	nHd += 8*len(wr.headerEdges) // headerEdges size
	nHd += 8*len(wr.dataEdges) // dataEdges size

	headerOffset := int64(nHd)
	dataOffset := int64(nHd) + wr.headerEdges[len(wr.headerEdges) - 1]
	for i := range wr.headerEdges {
		wr.headerEdges[i] += headerOffset
		wr.dataEdges[i] += dataOffset
	}

	err = binary.Write(fp, wr.order, wr.methodFlags)
	if err != nil { return nil, err }
	err = binary.Write(fp, wr.order, wr.headerEdges)
	if err != nil { return nil, err}
	err = binary.Write(fp, wr.order, wr.dataEdges)
	if err != nil { return nil, err }

	// Write the  header and data
	bHeader := wr.header.Bytes()
	bData := wr.data.Bytes()

	_, err = fp.Write(bHeader)
	if err != nil { return bData[:0], err }

	_, err = fp.Write(bData)
	if err != nil { return bData[:0], err }

	return bData, nil
}

type FixedWidthHeader struct {
	// N and Ntot give the number of particles in the file and in the
	// total simulation, respectively.
	N, NTot int64
	// Span gives the dimensions of the slab of particles in the file.
	// Span[0], Span[1], and Span[2] are the x-, y-, and z- dimensions.
	Span [3]int64
	// IDOffset is the ID of the first particle in the file.
	IDOffset uint64
	// Z, OmegaM, OmegaL, H100, L, and Mass give the redshift, Omega_m,
	// Omega_Lambda, H0 / (100 km/s/Mpc), box width in comoving Mpc/h,
	// and particle mass in Msun/h.
	Z, OmegaM, OmegaL, H100, L, Mass float64
}

type Header struct {
	FixedWidthHeader
	// OriginalHeader is the original header of the one of the simulation 
	OriginalHeader []byte
	// Names gives the names of all the variables stored in the file.
	// Types give the types of these variables. "u32"/"u64" give 32-bit and
	// 64-bit unisghned integers, respectively, anf "f32"/"f64" give 32-bit
	// and 64-bit floats, respectively.
	Names, Types []string
}

func convertSnapioHeader(
	snapioHeader snapio.Header, span [3]int64, idOffset uint64,
) *Header {
	n := span[0]*span[1]*span[2]
	return &Header{
		FixedWidthHeader{n, snapioHeader.NTot(), span, idOffset,
			snapioHeader.Z(), snapioHeader.OmegaM(),
			snapioHeader.OmegaL(), snapioHeader.H100(),
			snapioHeader.L(), snapioHeader.Mass()},
		snapioHeader.ToBytes(), []string{}, []string{},
	}	
}

func (hd *Header) read(f io.Reader, order binary.ByteOrder) error {
	err := binary.Read(f, order, &hd.FixedWidthHeader)
	if err != nil { return err }

	var nOHeader uint32
	if err := binary.Read(f, order, &nOHeader); err != nil { return err }
	hd.OriginalHeader = make([]byte, nOHeader)

	if _, err := f.Read(hd.OriginalHeader); err != nil { return err }

	var nFields uint32
	if err := binary.Read(f, order, &nFields); err != nil { return err }
	hd.Names, hd.Types = make([]string, nFields+1), make([]string, nFields+1)

	nNames := make([]uint32, nFields)
	if err := binary.Read(f, order, nNames); err != nil { return err }

	for i := 0; i < len(nNames); i++ {
		b := make([]byte, nNames[i])
		if _, err = f.Read(b); err != nil { return err }
		hd.Names[i] = string(b)
	}

	for i := 0; i < len(nNames); i++ {
		b := make([]byte, 3)
		if _, err = f.Read(b); err != nil { return err }
		hd.Types[i] = string(b)
	}

	hd.Names[nFields], hd.Types[nFields] = "id", "u64"
	return nil
}

func (hd *Header) write(f io.Writer, order binary.ByteOrder) (int, error) {
	n := 0
	err := binary.Write(f, order, &hd.FixedWidthHeader)
	if err != nil { return 0, err }
	n += int(unsafe.Sizeof(hd.FixedWidthHeader))

	nOHeader := uint32(len(hd.OriginalHeader))
	if err := binary.Write(f, order, nOHeader); err != nil { return 0, err }
	n += 4

	if _, err := f.Write(hd.OriginalHeader); err != nil { return 0, err }
	n += int(nOHeader)

	nFields := uint32(len(hd.Names))
	if err := binary.Write(f, order, &nFields); err != nil { return 0, err }
	n += 4

	nNames := make([]uint32, nFields)
	for i := range nNames { nNames[i] = uint32(len(hd.Names[i])) }
	if err := binary.Write(f, order, nNames); err != nil { return 0, err }
	n += len(nNames)*4

	for i := range hd.Names {
		b := []byte(hd.Names[i])
		if _, err = f.Write(b); err != nil { return 0, err }
		n += len(b)
	}

	for i := range hd.Types {
		b := []byte(hd.Types[i])
		if _, err = f.Write(b); err != nil { return 0, err }
		n += 3
	}

	return n, nil
}

// Reader handles the I/O and navigation asosociated with reading compressed
// fields from disk. Unlike Writer, it will need to be closed after use.
type Reader struct {
	Header
	fname string
	f *os.File
	order binary.ByteOrder
	headerEdges, dataEdges []int64
	methodFlags []MethodFlag
	buf *Buffer

	// I don't udnerstand why, but Go's zlib library crashes if I read directly
	// from disk, but not if I read into a bytes.Buffer and then decompress the
	// buffer.
	midBuf []byte
}

// NewReader creates a new Reader associated with the given gile and uses
// the given buffers to avoid unneccessary heap allocation.
func NewReader(
	fname string, buf *Buffer, midBuf []byte,
) (*Reader, error) {
	f, err := os.Open(fname)
	if err != nil { return nil, err }

	order, err := checkFile(fname, f)
	if err != nil { return nil, err }

	hd := &Header{ }
	if err := hd.read(f, order); err != nil { return nil, err }
	nFields := len(hd.Names) - 1

	rd := &Reader{
		*hd, fname, f, order, make([]int64, nFields+1),
		make([]int64, nFields+1), make([]MethodFlag, nFields), buf, midBuf,
	}

	// Read in navigation information
	if err := binary.Read(f, order, rd.methodFlags); err != nil {
		return nil, err
	}
	if err := binary.Read(f, order, rd.headerEdges); err != nil {
		return nil, err
	}
	if err := binary.Read(f, order, rd.dataEdges); err != nil {
		return nil, err
	}
	
	return rd, err
}

// ReadField reads a field from the reader using the given method. (Note: use
// Names() to find these.)
//
// NOTE: ReadField uses the array space in Buffer to allocate the Field. If you
// want to call ReadField again, YOU WILL NEED TO COPY THE DATA OUT OF THE FIELD
// and into your own locally-allocated array or you could lose it.
func (rd *Reader) ReadField(name string) (particles.Field, error) {
	if name == "id" { return rd.readID() }

	i := findString(rd.Names, name)
	if i == -1 { 
		return nil, fmt.Errorf("The field '%s' is not in the compressed " +
			"file %s. It conly contains the fields %s.",
			name, rd.fname, rd.Names)
	}

	headerOffset, dataOffset := rd.headerEdges[i], rd.dataEdges[i]

	_, err := rd.f.Seek(headerOffset, 0)
	if err != nil { return nil, err }

	// Select the method used
	method := selectMethod(rd.methodFlags[i])

	err = method.ReadInfo(rd.order, rd.f)
	if err != nil { return nil, err}

	_, err = rd.f.Seek(dataOffset, 0)
	if err != nil { return nil, err }

	// Some trickery due to the way Go's zlib library handles reading from
	// disk. I still don't understand why direct disk reads fail...
	// But this does have another benefit: it prevents the disk from being
	// locked while zlib is doing slow calculations.
	n := rd.dataEdges[i+1] - rd.dataEdges[i]
	rd.midBuf = resizeBytes(rd.midBuf, int(n))
	_, err = rd.f.Read(rd.midBuf)
	if err != nil { return nil, err }

	midBuf := bytes.NewBuffer(rd.midBuf)
	return method.Decompress(rd.buf, midBuf, name)
}

func (rd *Reader) readID() (particles.Field, error) {
	rd.buf.Resize(int(rd.N))
	ids := rd.buf.u64
	for i := range ids {
		ids[i] = uint64(i) + rd.IDOffset
	}

	return particles.NewUint64("id", ids), nil
}

// Close closes the files associated with the Reader.
func (rd *Reader) Close() {
	rd.f.Close()
}

// ReuseMidBuf returns the midBuf used by the Reader so that it can be used by
// a later reader without excess heap allocation.  
func (rd *Reader) ReuseMidBuf() []byte {
	return rd.midBuf[:0]
}

func selectMethod(flag MethodFlag) Method {
	switch flag {
	case LagrangianDeltaFlag: return &LagrangianDelta{ }
	default:
		panic(fmt.Sprintf("The method flag %d isn't recognized by " + 
			"Reader.ReadField. This is almost certianly an internal error, " + 
			"where there's some sort of file offset that's being read " + 
			"incorrectly. It could also be that your local version of Guppy " + 
			"is older than the version that made this file and there's a bug " + 
			"in Guppy's code that checks for this.", flag))
	}
}

// findString returns the index of the first instance of target in x and -1 if
// target isn't in x. 
func findString(x []string, target string) int {
	for i := range x {
		if x[i] == target { return i }
	}
	return -1
}

// checkFile reads in the file's magic number and version number and makes
// sure that guppy can actually read it. If it can, the byte order is returned.
// Otherwise an error is returned.
func checkFile(fname string, f *os.File) (binary.ByteOrder, error) {
	var magicNumber, version uint32

	// Read the magic number and check that this is actually a guppy file.
	order := binary.ByteOrder(binary.LittleEndian)
	err := binary.Read(f, order, &magicNumber)
	if err != nil { return nil, err }

	switch magicNumber {
	case MagicNumber:
	case ReverseMagicNumber: order = binary.BigEndian
	default:
		return order, fmt.Errorf("%s is not a guppy files. All guppy files " +
			"begin with either the 32-bit integer %x or %x. This file begins " +
			"with %x.", fname, MagicNumber, ReverseMagicNumber, magicNumber)
	}

	// Check the version.
	err = binary.Read(f, order, &version)
	if version > Version {
		return order, fmt.Errorf("The file %s was created with guppy version " + 
			"%d, but are trying to read it with guppy version %d. This means " +
			"that the file contains features which weren't implemented at " +
			"the time your code was written. You can download the latest " + 
			"version of guppy at github.com/phil-mansfield/guppy.", fname, 
			version, Version,
		)
	}

	return order, nil
}