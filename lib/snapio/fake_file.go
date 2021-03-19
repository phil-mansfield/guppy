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
	nTot int
}

// FakeFile is a object that implements the File interface for testing purposes,
// but can be initialized directly from arrays. See the File interface for
// method documentation.
type FakeFile struct {
	id, x io.Reader
	n, nTot int
}

// Type assertion
var (
	_ Header = &FakeFileHeader{ }
	_ File = &FakeFile{ }
)


// NewFakeFile creates a new FakeFile with the given IDs and position vectors.
// The snapshot has nTot files across all files and the byte order is given by
// order. The box will be 100 cMpc/h on a side and has z=0, Om = 0.27, and
// h100 = 0.7
func NewFakeFile(
	id []uint32, x [][3]float32,
	nTot int, order binary.ByteOrder,
) *FakeFile {
	n := len(id)
	return &FakeFile{
		arrayToReader(id, order), arrayToReader(x, order), n, nTot,
	}
}

func (f *FakeFile) ReadHeader() Header { return &FakeFileHeader{ f.nTot } }

func (f *FakeFile) Read(name string, buf *Buffer) error {
	switch name {
	case "id": return buf.read(f.id, "id", f.n)
	case "x": return buf.read(f.x, "x", f.n)
	default:
		return fmt.Errorf("FakeFileType only supports the variables 'x' and 'id', not '%s'", name)
	}
}

func (f *FakeFileHeader) ToBytes() []byte{ return []byte{4, 8, 15, 16, 23, 46} }
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
