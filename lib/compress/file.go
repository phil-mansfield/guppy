package compress

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/phil-mansfield/guppy/lib/particles"
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
	fname string
	buf *Buffer
	order binary.ByteOrder
	names []string
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
	fname string, buf *Buffer, b []byte, order binary.ByteOrder,
) (*Writer) {
	header := bytes.NewBuffer([]byte{ })
	data := bytes.NewBuffer(b) 
	return &Writer{
		fname, buf, order,
		[]string{}, []uint32{},
		[]int64{0}, []int64{0},
		header, data,
	}
}

// Add field adds a new field to the file which will be compressed with
// a given method.
func (wr *Writer) AddField(field particles.Field, method Method) error {
	err := method.WriteInfo(wr.header)
	if err != nil { return err }

	err = method.Compress(field, wr.buf, wr.data)
	if err != nil { return err }

	wr.headerEdges = append(wr.headerEdges, int64(wr.header.Len()))
	wr.dataEdges = append(wr.dataEdges, int64(wr.data.Len()))
	wr.methodFlags = append(wr.methodFlags, uint32(method.MethodFlag()))
	wr.names = append(wr.names, field.Name())

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

	// Write field names (needed for navigation).
	nFields := uint32(len(wr.dataEdges) - 1)

	err = binary.Write(fp, wr.order, nFields)
	if err != nil { return nil, err }
	nHd += 4

	nNames := make([]uint32, len(wr.names))
	for i := range nNames { nNames[i] = uint32(len(wr.names[i])) }

	err = binary.Write(fp, wr.order, nNames)
	if err != nil { return nil, err }
	nHd += 4*len(nNames)

	for i := range wr.names {
		n, err := fp.Write([]byte(wr.names[i]))
		if err != nil { return nil, err }
		nHd += n
	}

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

// Reader handles the I/O and navigation asosociated with reading compressed
// fields from disk. Unlike Writer, it will need to be closed after use.
type Reader struct {
	fname string
	f *os.File
	order binary.ByteOrder
	headerEdges, dataEdges []int64
	names []string
	methodFlags []MethodFlag
	buf *Buffer
}

// NewReader creates a new Reader associated with the given gile and uses
// the given buffer to avoid unneccessary heap allocation.
func NewReader(fname string, buf *Buffer) (*Reader, error) {
	f, err := os.Open(fname)
	if err != nil { return nil, err }

	order, err := checkFile(fname, f)
	if err != nil { return nil, err }

	var nFields uint64
	err = binary.Read(f, order, &nFields)

	rd := &Reader{
		fname, f, order, make([]int64, nFields), make([]int64, nFields),
		make([]string, nFields), make([]MethodFlag, nFields), buf,
	}
	nNames := make([]uint32, nFields)

	// Read in the field names
	err = binary.Read(f, order, nNames)
	if err != nil { return nil, err}

	for i := range rd.names {
		b := make([]byte, nNames[i])
		_, err = f.Read(b)
		if err != nil { return nil, err }
		rd.names[i] = string(b)
	}

	// Read in navigation information
	err = binary.Read(f, order, rd.methodFlags)
	if err != nil { return nil, err }
	err = binary.Read(f, order, rd.headerEdges)
	if err != nil { return nil, err }
	err = binary.Read(f, order, rd.dataEdges)
	if err != nil { return nil, err }

	return rd, err
}

func (rd *Reader) Names() []string { return rd.names }
func (rd *Reader) MethodFlags() []MethodFlag {return rd.methodFlags }

// ReadField reads a field from the reader using the given method. (Note: use
// Names() and MethodFlags() to find these.)
func (rd *Reader) ReadField(
	name string, method Method,
) (particles.Field, error) {
	
	i := findString(rd.names, name)
	if i == -1 { 
		return nil, fmt.Errorf("The field '%s' is not in the compressed " +
			"file %s. It conly contains the fields %s.",
			name, rd.fname, rd.names)
	}

	headerOffset, dataOffset := rd.headerEdges[i], rd.dataEdges[i]

	_, err := rd.f.Seek(headerOffset, 0)
	if err != nil { return nil, err }

	err = method.ReadInfo(rd.order, rd.f)
	if err != nil { return nil, err}

	_, err = rd.f.Seek(dataOffset, 0)
	if err != nil { return nil, err }

	return method.Decompress(rd.buf, rd.f)
}

// Close closes the files associated with the Reader.
func (rd *Reader) Close() {
	rd.f.Close()
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
	var magicNumber, version uint64

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