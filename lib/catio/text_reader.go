package catio

import (
	"bytes"
	"io"
	"runtime"
)

type textReader struct {
	rd io.ReadSeeker
	config TextConfig
	size int
	blockStarts []int
	blockEnds []int
	buf []byte
}

// newTextReader creates a new textReader associated with the I/O stream rd,
// which contains size bytes. An optional config file can be provided,
// otherwise DefaultConfig will be used.
func newTextReader(
	rd io.ReadSeeker, size int, config ...TextConfig,
) *textReader {
	reader := &textReader{ config: DefaultConfig, size: size, rd: rd }
	if len(config) > 0 { reader.config = config[0] }
	

	// Figure out how many blocks are in the file.
	blocks := 1 + size / reader.config.MaxBlockSize
	if (blocks - 1)*reader.config.MaxBlockSize == size { blocks-- }

	reader.blockStarts = make([]int, blocks)
	reader.blockEnds = make([]int, blocks)

	// Find the start of each block.
	buf := make([]byte, reader.config.MaxLineSize)
	for i := 0; i < blocks; i++ {
		reader.blockStarts[i] = reader.blockStart(i, buf)
	}
	
	// Find the end of each block.
	for i := 0; i < len(reader.blockEnds) - 1; i++ {
		reader.blockEnds[i] = reader.blockStarts[i+1]
	}

	reader.blockEnds[blocks - 1] = size

	// initialize byte buffer
	maxSize := 0
	for i := range reader.blockStarts {
		size := reader.blockEnds[i] - reader.blockStarts[i]
		if size > maxSize { maxSize = size }
	}
	reader.buf = make([]byte, maxSize)
	
	return reader
}

// blockStart returns the index of the starting byte of the specified byte. It
// requires a buffer that is large enough to read any line of the catalogue
// file.
func (t *textReader) blockStart(block int, buf []byte) int {
	if block == 0 { return 0 }

	// starting and ending indices of the line surrounding the block break
	lineEnd := block * t.config.MaxBlockSize
	if lineEnd > t.size { lineEnd = t.size }
	lineStart := lineEnd - len(buf)

	// Find the start of the line...
	_, err := t.rd.Seek(int64(lineStart), 0)
	if err != nil { panic(err.Error()) }
	
	// ...and read it
	_, err = io.ReadAtLeast(t.rd, buf, len(buf))
	if err != nil { panic(err.Error()) }

	// Find the first line break
	idx := bytes.IndexByte(buf, '\n')
	if idx == -1 { panic("Can't find newline in line.") }

	return idx + lineStart
}

// columnIndices converts the generic columns variable into integer indices.
// If columns is []int, it returns them, if columns is []string, it looks up
// the corresponding ints.
func (t *textReader) columnIndices(columns interface{}) []int {
	if intCols, ok := columns.([]int); ok {
		return intCols
	} else if strCols, ok := columns.([]string); ok {
		idxs := make([]int, len(strCols))
		for i := range strCols {
			if idx, ok := t.config.ColumnNames[strCols[i]]; ok {
				idxs[i] = idx
			}
		}
	} 
	panic("Columns argument must be []int or []string.")
}

// ReadInts reads the specified columns from every block in the file,
// interprets them as ints and concetenates them together.
func (t *textReader) ReadInts(
	columns interface{}, bufs ...[][]int,
) [][]int {
	idx := t.columnIndices(columns)
	buf := make([][]int, len(idx))
	if len(bufs) > 0 { buf = bufs[0] }

	start := 0
	for i := 0; i < t.Blocks(); i++ {
		skip := t.config.SkipLines
		if i > 0 { skip = 0 }
		buf, start = t.bufferedReadInts(idx, i, buf, start, skip)
	}
	
	clipIntBuffers(buf, start)

	return buf
}

func (t *textReader) ReadFloat64s(
	columns interface{}, bufs ...[][]float64,
) [][]float64 {
	idx := t.columnIndices(columns)
	buf := make([][]float64, len(idx))
	if len(bufs) > 0 { buf = bufs[0] }

	start := 0
	for i := 0; i < t.Blocks(); i++ {
		skip := t.config.SkipLines
		if i > 0 { skip = 0 }
		buf, start = t.bufferedReadFloat64s(idx, i, buf, start, skip)
	}
	
	clipFloat64Buffers(buf, start)

	return buf
}

func (t *textReader) ReadFloat32s(
	columns interface{}, bufs ...[][]float32,
) [][]float32 {
	idx := t.columnIndices(columns)
	buf := make([][]float32, len(idx))
	if len(bufs) > 0 { buf = bufs[0] }

	start := 0
	for i := 0; i < t.Blocks(); i++ {
		skip := t.config.SkipLines
		if i > 0 { skip = 0 }
		buf, start = t.bufferedReadFloat32s(idx, i, buf, start, skip)
	}
	
	clipFloat32Buffers(buf, start)

	return buf
}

func (t *textReader) Blocks() int {
	return len(t.blockStarts)
}

// ReadIntBlock reads the specified columns from the given block as ints.
func (t *textReader) ReadIntBlock(
	columns interface{}, i int, bufs ...[][]int,
) [][]int {
	idx := t.columnIndices(columns)
	buf := make([][]int, len(idx))
	if len(bufs) > 0 { buf = bufs[0] }

	skip := t.config.SkipLines
	if i > 0 { skip = 0 }
	buf, end := t.bufferedReadInts(idx, i, buf, 0, skip)

	clipIntBuffers(buf, end)

	return buf
}

func (t *textReader) ReadFloat64Block(
	columns interface{}, i int, bufs ...[][]float64,
) [][]float64 {
	idx := t.columnIndices(columns)
	buf := make([][]float64, len(idx))
	if len(bufs) > 0 { buf = bufs[0] }

	skip := t.config.SkipLines
	if i > 0 { skip = 0 }
	buf, end := t.bufferedReadFloat64s(idx, i, buf, 0, skip)

	clipFloat64Buffers(buf, end)

	return buf
}

func (t *textReader) ReadFloat32Block(
	columns interface{}, i int, bufs ...[][]float32,
) [][]float32 {
	idx := t.columnIndices(columns)
	buf := make([][]float32, len(idx))
	if len(bufs) > 0 { buf = bufs[0] }

	skip := t.config.SkipLines
	if i > 0 { skip = 0 }
	buf, end := t.bufferedReadFloat32s(idx, i, buf, 0, skip)

	clipFloat32Buffers(buf, end)

	return buf
}

func (t *textReader) bufferedReadInts(
	idxs []int, i int, bufs [][]int, start, skip int,
) (outBuf [][]int, end int) {
	runtime.GC()

	// Read raw bytes.
	n := t.blockEnds[i] - t.blockStarts[i]
	_, err := t.rd.Seek(int64(t.blockStarts[i]), 0)
	if err != nil { panic(err.Error()) }
	_, err = io.ReadAtLeast(t.rd, t.buf, n)
	if err != nil { panic(err.Error()) }

	// Separate and clean lines
	lines, nComm := split(t.buf[:n], '\n', t.config.Comment)
	lines = lines[skip:]
	lines = uncomment(lines, t.config.Comment, nComm)
	lines = trim(lines, t.config.Separator)

	// Increase buffer size if needed
	for i := range bufs { bufs[i] = bufs[i][:cap(bufs[i])] }
	avail := len(bufs[0]) - start
	if avail < len(lines) {
		for i := range bufs {
			bufs[i] = append(bufs[i], make([]int, len(lines) - avail)...)
		}
	}

	// Tranform into a parse-friendly format.
	parseBufs := make([][]int, len(bufs))
	for i := range parseBufs {
		parseBufs[i] = bufs[i][start: start + len(lines)]
	}

	// Parse!
	err = parseInts(lines, t.config.Separator, idxs, parseBufs)
	if err != nil { panic(err.Error()) }
	
	return bufs, start + len(lines)
}

func (t *textReader) bufferedReadFloat64s(
	idxs []int, i int, bufs [][]float64, start, skip int,
) (outBuf [][]float64, end int) {
	runtime.GC()

	// Read raw bytes.
	n := t.blockEnds[i] - t.blockStarts[i]
	_, err := t.rd.Seek(int64(t.blockStarts[i]), 0)
	if err != nil { panic(err.Error()) }
	_, err = io.ReadAtLeast(t.rd, t.buf, n)
	if err != nil { panic(err.Error()) }

	// Separate and clean lines
	lines, nComm := split(t.buf[:n], '\n', t.config.Comment)
	lines = lines[skip:]
	lines = uncomment(lines, t.config.Comment, nComm)
	lines = trim(lines, t.config.Separator)

	// Increase buffer size if needed
	for i := range bufs { bufs[i] = bufs[i][:cap(bufs[i])] }
	avail := len(bufs[0]) - start
	if avail < len(lines) {
		for i := range bufs {
			bufs[i] = append(bufs[i], make([]float64, len(lines) - avail)...)
		}
	}

	// Tranform into a parse-friendly format.
	parseBufs := make([][]float64, len(bufs))
	for i := range parseBufs {
		parseBufs[i] = bufs[i][start: start + len(lines)]
	}

	// Parse!
	err = parseFloat64s(lines, t.config.Separator, idxs, parseBufs)
	if err != nil { panic(err.Error()) }
	
	return bufs, start + len(lines)
}

func (t *textReader) bufferedReadFloat32s(
	idxs []int, i int, bufs [][]float32, start, skip int,
) (outBuf [][]float32, end int) {
	runtime.GC()

	// Read raw bytes.
	n := t.blockEnds[i] - t.blockStarts[i]
	_, err := t.rd.Seek(int64(t.blockStarts[i]), 0)
	if err != nil { panic(err.Error()) }
	_, err = io.ReadAtLeast(t.rd, t.buf, n)
	if err != nil { panic(err.Error()) }

	// Separate and clean lines
	lines, nComm := split(t.buf[:n], '\n', t.config.Comment)
	lines = lines[skip:]
	lines = uncomment(lines, t.config.Comment, nComm)
	lines = trim(lines, t.config.Separator)

	// Increase buffer size if needed
	for i := range bufs { bufs[i] = bufs[i][:cap(bufs[i])] }
	avail := len(bufs[0]) - start
	if avail < len(lines) {
		for i := range bufs {
			bufs[i] = append(bufs[i], make([]float32, len(lines) - avail)...)
		}
	}

	// Tranform into a parse-friendly format.
	parseBufs := make([][]float32, len(bufs))
	for i := range parseBufs {
		parseBufs[i] = bufs[i][start: start + len(lines)]
	}

	// Parse!
	err = parseFloat32s(lines, t.config.Separator, idxs, parseBufs)
	if err != nil { panic(err.Error()) }
	
	return bufs, start + len(lines)
}

// clipIntBuffers slices all the buffers in bufs so that they are of length n.
func clipIntBuffers(bufs [][]int, n int) {
	for i := range bufs {
		bufs[i] = bufs[i][:n]
	}
}

func clipFloat64Buffers(bufs [][]float64, n int) {
	for i := range bufs {
		bufs[i] = bufs[i][:n]
	}
}

func clipFloat32Buffers(bufs [][]float32, n int) {
	for i := range bufs {
		bufs[i] = bufs[i][:n]
	}
}
