/*package format handles Guppy's miniature formatting languages for snapshot
file, e.g:

   Input = "snapdir{%03d,snapshot}/snap{%03d,snapshot}.{%d,0..512}" 
   Output = "snapdir{%04d,snapshot}/snap{%04d,snapshot}.{%d,output}.gup"
   Snapshots = 0..100 - 63

The exact rules are as follows:
File format strings are a combination of fixed text and variables. Fixed text is
always the same, and variables can change from file to file. Variables are
written as {verb,rule}. "verb" is a printf() verb (e.g. %03d) that specifies
how the variable should be printed. "rule" is text that specifies what values
the variable should take on. There are currently three rules:

  "snapshot" - The variable is equal to the currently-analysed snapshot.
  "output" - The variable ranges over Guppy output file indices.
  sequence format - The variable ranges over a user-specified range.

Sequence formats are a generic way to specify non-contiguous sequences of
natural numbers. They consist of a series of n tokens separated by "+" or "-".
Each token can be either a number or two numbers separted by "..". E.g.:

  100
  0..100
  0..10 + 100
  0..100 - 63 - 10..20

These strings build up sequences of numbers by adding/removing individual
numbers and contiguous sequences. For example, 0 through 10 would be 0..10,
1, 2, 3, 15, 16, 17 could be written as  1..17 - 4..13. This is useful for
skipping corrupted snapshots or specifying a subset of snapshots/files.

All spaces around "-", "+", and "," symbols are ignored.
*/
package format

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	
	g_error "github.com/phil-mansfield/guppy/lib/error"
)

const (
	// Any expanded formats which would have more than BigNumber elements are
	// assumed to be bugs.
	BigNumber = 1<<20
)

// ExpandSequenceFormat expands a sequence format string into a sorted sequence
// of integers.
func ExpandSequenceFormat(format string) ([]int, error) {
	// Parse and error-check the format string.
	tok, err := tokeniseSequenceFormat(format)
	if err != nil { return nil, err }
	adds, subs, err := addsSubsSequenceFormat(tok)
	if err != nil { return nil, err }

	// Add numbers to the sequence.
	m := map[int]int{ }
	for i := range adds {
		ns := parseSequenceFormatToken(adds[i])
		for _, n := range ns {
			if _, ok := m[n]; ok {
				return nil, fmt.Errorf("The number %d is added more than once.", n)
			}
			m[n] = n
		}
	}

	// Remove numbers from the sequence.
	for i := range subs {
		ns := parseSequenceFormatToken(subs[i])
		for _, n := range ns {
			if _, ok := m[n]; !ok {
				return nil, fmt.Errorf("The number %d is removed more times than it was inserted.", n)
			}
			delete(m, n)
		}
	}
	
	if len(m) > BigNumber {
		return nil, fmt.Errorf("This sequence would have %d elements, which is almost certianly a bug.", len(m))
	}

	// Convert to a sorted array of integers.
	out := []int{ }
	for n := range m { out = append(out, n) }
	sort.Ints(out)
	
	return out, nil
}

// tokeniseSequenceFormat tokenizes a SeqeunceFormat string. This means that
// is separates all the 
func tokeniseSequenceFormat(format string) ([]string, error) {
	// Make sure all operators are separated by spaces.
	formatClean := strings.ReplaceAll(format, "+", " + ")
	formatClean = strings.ReplaceAll(formatClean, "-", " - ")

	// Tokenize and remove empty tokens.
	tokRaw := strings.Split(formatClean, " ")
	tok := []string{ }
	for i := range tokRaw {
		tokRaw[i] = strings.Trim(tokRaw[i], " ")
		if len(tokRaw[i]) > 0 {
			tok = append(tok, tokRaw[i])
		}
	}
	
	if len(tok) == 0 {
		return nil, fmt.Errorf("The format string is empty.")
	}
	return tok, nil
}

func addsSubsSequenceFormat(tok []string) (adds, subs []string, err error) {
	if len(tok) == 0 {
		return nil, nil, fmt.Errorf("Format string is empty")
	}

	
	// Handle the case where the starting "+" is dropped.
	adds, subs = []string{}, []string{}
	var start int
	if tok[0] == "+" || tok[0] == "-" {
		start = 0
	} else {
		if err := isSequenceFormatToken(tok[0]); err != nil {
			return nil, nil, fmt.Errorf(
				"Element number %d, '%s', cannot be parsed because %s",
				1, tok[0], err.Error(),
			)
		}
		
		adds = append(adds, tok[0])
		start = 1
	}

	for i := start; i < len(tok); i += 2 {
		if tok[i] != "-" && tok[i] != "+" {
			return nil, nil, fmt.Errorf(
				"Element number %d, '%s', should be a '-' or '+', but isn't.",
				i+1, tok[i])
		}

		if i + 1 >= len(tok) {
			return nil, nil, fmt.Errorf(
				"The format string ends in a trailing '%s'", tok[i],
			)
		}

		if err := isSequenceFormatToken(tok[i+1]); err != nil {
			return nil, nil, fmt.Errorf(
				"Element number %d, '%s', cannot be parsed because %s",
				i+2, tok[i+1], err.Error(),
			)
		}
		
		if tok[i] == "+" {
			adds = append(adds, tok[i+1])
		} else {
			subs = append(subs, tok[i+1])
		}
	}

	return adds, subs, nil
}

// isSequenceFormatToken returns a nil error is tok is a valid token for
// a sequence format and an error describing the problem otherwise. The error
// message assumes it is printed after a trailing "beacause"
func isSequenceFormatToken(tok string) error {
	if len(tok) == 0 {
		return fmt.Errorf("the format string is empty.")
	}
	
	bounds := strings.Split(tok, "..")

	switch len(bounds) {
	case 1:
		_, err := strconv.Atoi(bounds[0])
		if err != nil {
			return fmt.Errorf("'%s' is not an integer.", bounds[0])
		}
		return nil
	case 2:
		start, err1 := strconv.Atoi(bounds[0])
		if err1 != nil {
			return fmt.Errorf("'%s' is not an integer.", bounds[0])
		}
		end, err2 := strconv.Atoi(bounds[1])
		if err2 != nil {
			return fmt.Errorf("'%s' is not an integer.", bounds[1])
		}
		if end < start {
			return fmt.Errorf("lower bound %d is larger than upper bound %d.",
				start, end)
		}
		
		return nil
	}
	return fmt.Errorf("it has more than one '..'.")
}

// parseSeqeunceFormatToken parses a single token in a seqeunce format stirng
// and returns the corresponding array of numbers. This function assumes that
// the tests in isSequenceFormatToken have already been run and thus does no
// error checking. This makes sense to do because the calling funciton has
// already removed location information from these tokens, so the error
// message would be less informative.
func parseSequenceFormatToken(tok string) []int {
	bounds := strings.Split(tok, "..")

	switch len(bounds) {
	case 1:
		n, _ := strconv.Atoi(tok)
		return []int{ n }
	case 2:
		start, _ := strconv.Atoi(bounds[0])
		end, _ := strconv.Atoi(bounds[1])
		out := []int{ }
		for n := start; n <= end; n++ {
			out = append(out, n)
		}

		return out
	}

	g_error.Internal(
		"Invalid sequence format token, '%s', passed isSeqeunceFormatToken()",
		tok,
	)
	return nil
}

// ExpandSnapshotFormat expands the format string specifying the snapshots
// Guppy will analyse. This format string takes the same form as the general
// ExpandSequeneceFormat format string.
func ExpandSnapshotFormat(format string) []int {
	snaps, err := ExpandSequenceFormat(format)
	if err != nil {
		g_error.External("The Snapshots format string, '%s' is not valid. %s",
			format, err.Error(),
		)
	}

	return snaps
}

func ExpandFileFormat(format string, snapshot, output int) {
	starts, ends := fileFormatStartsEnds(format)
	comp := NewFileFormatComponents(format, starts, ends)
	_ = comp
	panic("NYI")
}

// fileFormatStartsEnds returns the indices of the beginning and end of each
// format variable.
func fileFormatStartsEnds(format string) (starts, ends []int) {
	starts, ends = []int{ }, []int{ }
	nestedLevel := 0

	ending := "Make sure variables in file formats are enclosed in matching { ... } pairs."
	
	for i := range format {
		if format[i] == '{' {
			nestedLevel++
			starts = append(starts, i)
		} else if format[i] == '}' {
			nestedLevel--
			ends = append(ends, i+1)
		}

		if nestedLevel > 1 {
			end := len(starts) - 1
			g_error.External("The file format '%s' has nested '{' characters, making it invalid. These '{'s are at indices %d and %d. " + ending,
				format, starts[end], starts[end - 1])
		} else if nestedLevel < 0 {
			end := len(ends) - 1
			g_error.External("The file format '%s' has a '}' that doesn't come after a '{' character, making it invalid. This '}' is at index %d. " + ending,
				format, ends[end],
			)
		}
	}

	if len(ends) != len(starts) {
		end := len(starts) - 1
		g_error.External("The file format '%s' has a '{' without a matching '}', making it invalid. This '{' is at index %d. " + ending, format, starts[end])
	}

	return starts, ends
}

type FileFormatComponents struct {
	Separators []string
	Vars []string
	FormatVerbs map[string]string
	Arguments map[string]string
}

func NewFileFormatComponents(
	format string, starts, ends []int,
) *FileFormatComponents {
	comp := &FileFormatComponents{ }

	sepStart := 0
	for i := range starts {
		comp.Separators = append(comp.Separators, format[sepStart: starts[i]])

		v := format[starts[i]+1: ends[i]-1]

		//base := fmt.Sprintf("The file format '%s' has an invalid variable, '%s'. Variables should contain a formatting 'verb' (e.g. '%%d', '%%03d', etc.), a comma, and an argument giving the values that variable takes on (e.g. '0..511', 'snapshot', etc.)", format, )

		_ = v //base
		/*
		tok := strings.Split(v, ",")
		if len(tok) != 2 {
			g_error.External(base + )
		}
*/
	}

	panic("NYI")
	return comp
}
