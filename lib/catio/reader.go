package catio

import (
	"os"
	"bytes"
	"io/ioutil"
)

// TextConfig contains information neccessary for parsing halo catalogues.
type TextConfig struct {
	Separator byte // Character used to separated fields
	Comment byte // Character used to start comments.
	SkipLines int // Number of lines to skip at the start of file.
	ColumnNames map[string]int // Map from column names to 
	MaxBlockSize int // Largest amount of text you want to read at one time.
	MaxLineSize int // Largest possible line size.
}

// DefaultConfig is a TextConfig instance which can read arbitrary Rockstar
// files.
var DefaultConfig = TextConfig{
	Separator: ' ',
	Comment: '#',
	SkipLines: 0,
	ColumnNames: map[string]int{},
	
	MaxBlockSize: 10 * 1<<30,
	MaxLineSize: 1<<20,
}

// Reader allows the user to access data fields in a halo catalog of different
// types, potentially in blocks.
type Reader interface {
	// Read* methods reads data across all blocks simultaneously. Optional
	// buffers may be provided if you're worried about allocation.
	ReadInts(columns interface{}, bufs ...[][]int) [][]int
	ReadFloat64s(columns interface{}, bufs ...[][]float64) [][]float64
	ReadFloat32s(columns interface{}, bufs ...[][]float32) [][]float32
	
	// Blocks returns the number of blocks in the halo file.
	Blocks() int
	
	// Read*Int reads data associated with block i. Optional buffers may be
	// provided if you're worried about allocation.
	ReadIntBlock(columns interface{}, i int, bufs ...[][]int) [][]int
	ReadFloat64Block(columns interface{}, i int,
		bufs ...[][]float64) [][]float64
	ReadFloat32Block(columns interface{}, i int,
		bufs ...[][]float32) [][]float32
}

// TextFile creates a Reader for standard text-based halo file.
func TextFile(fname string, config ...TextConfig) Reader {
    f, err := os.Open(fname)
	if err != nil { panic(err.Error()) }
    info, err := f.Stat()
	if err != nil { panic(err.Error()) }

	return newTextReader(f, int(info.Size()), config...)
}

// Text creates a Reader for a block of text.
func Text(text []byte, config ...TextConfig) Reader {
	return newTextReader(bytes.NewReader(text), len(text), config...)
}

// Stdin creates a Reader for the text currently in stdin.
func Stdin(config ...TextConfig) Reader {
    text, err  := ioutil.ReadAll(os.Stdin)
    if err != nil { panic(err.Error()) }
	return Text(text, config...)
}
