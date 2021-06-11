package catio

import (
	"fmt"
	"strconv"
	"bytes"
	"runtime"
)

// split splits a byte splice at each separating flag. Faster than
// bytes.Split() because slicing is used instead of allocations and because
// only one separator is used.
//
// Some of the calculations associated with uncommenting are done here for a
// slight performance boost.
func split(data []byte, sep, comm byte) (lines [][]byte, nComm int) {
	n, nComm := 0, 0
	for _, c := range data {
		if c == sep { n++ }
		if c == comm { nComm++ }
	}

	tokens := make([][]byte, n+1)

	idx := 0
	for j := 0; j < n; j++ {
		data = data[idx:]
		idx = bytes.IndexByte(data, sep)
		tokens[j] = data[:idx]
		idx++
	}
	tokens[n] = data[idx:]

	return tokens, nComm
}

// uncomment removes file comments  in the form of "data # comment". Optimized
// for the common case where comments are rare and at the start of the file.
func uncomment(lines [][]byte, comm byte, nComm int) [][]byte {
	if nComm == 0 { return lines }

	for i, line := range lines {
		commentStart := bytes.IndexByte(line, comm)
		if commentStart == -1 {
			continue
		}

		lines[i] = line[:commentStart]

		n := 1
		for _, c := range line[commentStart+1:] {
			if c == comm { n++ }
		}

		nComm -= n
		if nComm == 0 { return lines }
	}

	return lines
}

// trim removes empty lines.
func trim(lines [][]byte, sep byte) [][]byte {
	j := 0

	LineLoop:
	for i, line := range lines {
		for _, c := range line {
			if c != sep {
				lines[j] = lines[i]
				j++
				continue LineLoop
			}
		}
	}

	return lines[:j]
}

func parseInts(lines [][]byte, sep byte, idxs []int, out [][]int) error {
	if len(lines) == 0 || len(idxs) == 0 { return nil }
	buf := make([][]byte, len(bytes.Fields(lines[0])))

	maxCol := -1
	for _, i := range idxs {
		if i > maxCol { maxCol = i }
	}

	if maxCol >= len(buf) {
		return fmt.Errorf(
			"Data has %d columns, but column %d was requested.",
			len(buf), maxCol,
		)
	}

	var err error
	for i, line := range lines {

		// Break line up into fields/words

		words := fields(line, sep, buf)
		if len(words) != len(buf) {
			return fmt.Errorf(
				"Data on line %d has %d columns, not %d.",
				i+1, len(words), len(buf),
			)
		}

		// Parse strings.

		for j := range idxs {
			out[j][i], err = strconv.Atoi(string(words[idxs[j]]))
			if err != nil { return err }
		}
	}

	return nil
}

func parseFloat64s(
	lines [][]byte, sep byte, idxs []int, out [][]float64,
) error {
	if len(lines) == 0 || len(idxs) == 0 { return nil }
	buf := make([][]byte, len(bytes.Fields(lines[0])))

	maxCol := -1
	for _, i := range idxs {
		if i > maxCol { maxCol = i }
	}

	if maxCol >= len(buf) {
		return fmt.Errorf(
			"Data has %d columns, but column %d was requested.",
			len(buf), maxCol,
		)
	}

	var err error
	for i, line := range lines {

		if i % 1000000 == 0 { runtime.GC() }

		// Break line up into fields/words

		words := fields(line, sep, buf)
		if len(words) != len(buf) {
			return fmt.Errorf(
				"Data on line %d has %d columns, not %d.",
				i+1, len(words), len(buf),
			)
		}

		// Parse strings.

		for j := range idxs {
			out[j][i], err = strconv.ParseFloat(string(words[idxs[j]]), 64)
			if err != nil { return err }
		}
	}

	return nil
}

func parseFloat32s(
	lines [][]byte, sep byte, idxs []int, out [][]float32,
) error {
	if len(lines) == 0 || len(idxs) == 0 { return nil }
	buf := make([][]byte, len(bytes.Fields(lines[0])))

	maxCol := -1
	for _, i := range idxs {
		if i > maxCol { maxCol = i }
	}

	if maxCol >= len(buf) {
		return fmt.Errorf(
			"Data has %d columns, but column %d was requested.",
			len(buf), maxCol,
		)
	}

	for i, line := range lines {

		// Break line up into fields/words

		words := fields(line, sep, buf)
		if len(words) != len(buf) {
			return fmt.Errorf(
				"Data on line %d has %d columns, not %d.",
				i+1, len(words), len(buf),
			)
		}

		// Parse strings.

		for j := range idxs {
			f, err := strconv.ParseFloat(string(words[idxs[j]]), 32)
			if err != nil { return err }
			out[j][i] = float32(f)
		}
	}

	return nil
}

// Optimized and buffered analog to the standard library's bytes.FieldsFunc()
// function.
func fields(data []byte, sep byte, buf [][]byte) [][]byte {
	n := 0
	inField := false
	for _, c := range data {
		wasInField := inField
		inField = sep != c
		if inField && !wasInField { n++ }
	}

	na := 0
	fieldStart := -1

	for i := 0; i < len(data) && na < n; i++ {
		c := data[i]

		if fieldStart < 0 && c != sep {
			fieldStart = i
			continue
		}

		if fieldStart >= 0 && c == sep {
			buf[na] = data[fieldStart: i]
			na++
			fieldStart = -1
		}
	}

	if fieldStart >= 0 {
		buf[na] = data[fieldStart: len(data)]
		na++
	}

	return buf[0:na]
}
