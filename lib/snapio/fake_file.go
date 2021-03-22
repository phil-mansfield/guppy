package snapio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// FakeFileHEader implements the Header interface for the purposes of giving
// FakeFile a valid ReaderHeader() return value. See the Header interface for
// method documentation.
type FakeFileHeader struct {
	names, types []string
	nTot int
}

// FakeFile is a object that implements the File interface for testing purposes,
// but can be initialized directly from arrays. See the File interface for
// method documentation.
type FakeFile struct {
	names, types []string
	readers []io.Reader
	n, nTot int
	order binary.ByteOrder
}

// Type assertion
var (
	_ Header = &FakeFileHeader{ }
	_ File = &FakeFile{ }
)


// NewFakeFile creates a new FakeFile with the given IDs and position vectors.
// The snapshot has nTot files across all files and the byte order is given by
// order. The box will be 100 cMpc/h on a side and has z=0, Om = 0.27, and
// h100 = 0.7.
func NewFakeFile(
	names []string, values []interface{},
	nTot int, order binary.ByteOrder,
) (*FakeFile, error) {
	n := -1
	readers := make([]io.Reader, len(names))
	types := make([]string, len(names))
	for i := range names {
		if names[i] == "id" { n = len(names[i]) }
		readers[i] = arrayToReader(values[i], order)

		switch values[i].(type) {
		case []uint32: types[i] = "u32"
		case []uint64: types[i] = "u64"
		case []float32: types[i] = "f32"
		case []float64: types[i] = "f64"
		case [][3]float32: types[i] = "v32"
		case [][3]float64: types[i] = "v64"
		default:
			return nil, fmt.Errorf(
				"Value %d in values is an unsupported type.", i,
			)
		}
	}
	
	return &FakeFile{
		names, types, readers, n, nTot, order,
	}, nil
}

func (f *FakeFile) ReadHeader() Header {
	return &FakeFileHeader{ f.names, f.types, f.nTot }
}

func (f *FakeFile) Read(name string, buf *Buffer) error {
	for i := range f.names {
		if name == f.names[i] {
			return buf.read(f.readers[i], name, f.n)
		}
	}
	return fmt.Errorf("There is no field '%s' in the file.", name)
}

func (f *FakeFileHeader) ToBytes() []byte{ return []byte{4, 8, 15, 16, 23, 46} }
func (f *FakeFileHeader) Names() []string { return f.names }
func (f *FakeFileHeader) Types() []string { return f.types }
func (f *FakeFileHeader) NTot() int { return f.nTot }
func (f *FakeFileHeader) Z() float64 { return 0.0 }
func (f *FakeFileHeader) OmegaM() float64 { return 0.27 }
func (f *FakeFileHeader) H100() float64 { return 0.70 }
func (f *FakeFileHeader) L() float64 { return 100.0 }

func arrayToReader(x interface{}, order binary.ByteOrder) io.Reader {
	buf := &bytes.Buffer{ }
	err := binary.Write(buf, order, x)
	if err != nil { panic(err.Error()) }

	rd := bytes.NewReader(buf.Bytes())
	return rd
}
