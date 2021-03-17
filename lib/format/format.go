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

var (
	LowerCase = "abcdefghijklmnopqrstuvwxyz"
	AllLetters = LowerCase + strings.ToUpper(LowerCase)
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

// addsSubsSequenceFormat takes a tokenised string and converts it to
// and adds array, representing all the tokens that should be inserted to the
// sequence and subs, all the tokens that should be removed from it.
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
// message assumes it is printed after a trailing "because".
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

// ExpandFormatString expands a format string into n unique versions. Named
// variables are given the values specified in vals. Sequence variables are
// expanded individually, meaning that n = \prod n_i, where n_i is the length
// of each sequence.
func ExpandFormatString(format string, vals map[string]int) ([]string, error) {
	// Toeknize components of the format string:
	starts, ends, err := startsEndsFormatString(format)
	if err != nil { return nil, err }
	text, vars := splitFormatString(format, starts, ends)
	
	// Parse each variable
	verb, rule := make([]string, len(vars)), make([]string, len(vars))
	isSeq := make([]bool, len(vars))
	for i := range vars {
		verb[i], rule[i], isSeq[i], err = splitFormatStringVar(vars[i], vals)
		if err != nil {
			return nil, fmt.Errorf("Could not parse variable %d, {%s}, because %s",
				i+1, vars[i], err.Error())
		}
	}

	expSeq, nTot := expandSeqValues(rule, isSeq)
	if nTot == 0 { nTot = 1 }
	
	// Construct the strings from the various rules and format verbs.
	out := make([]string, nTot)
	for i := range out {
		tok := []string{ }
		for j := range rule {
			tok = append(tok, text[j])
			if isSeq[j] {
				tok = append(tok, fmt.Sprintf(verb[j], expSeq[j][i]))
			} else {
				tok = append(tok, fmt.Sprintf(verb[j], vals[rule[j]]))
			}
		}
		
		tok = append(tok, text[len(text) - 1])
		
		out[i] = strings.Join(tok, "")
	}
	
	return out, nil
}


// startsEndsFormatString returns the indices of the beginning and end of each
// format variable. The end index is exclusive, making it useful for slicing
// the string.
func startsEndsFormatString(format string) (starts, ends []int, err error) {
	starts, ends = []int{ }, []int{ }
	nestedLevel := 0
	
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
			return nil, nil, fmt.Errorf("The format string '%s' has nested '{' characters, making it invalid. These two '{'s are characters %d and %d.",
				format, starts[end-1] + 1, starts[end] + 1)
		} else if nestedLevel < 0 {
			end := len(ends) - 1
			return nil, nil, fmt.Errorf("The format string '%s' has a '}' that doesn't come after a '{' character, making it invalid. This '}' is character %d.",
				format, ends[end],
			)
		}
	}

	if len(ends) != len(starts) {
		end := len(starts) - 1
		return nil, nil, fmt.Errorf("The file format '%s' has a '{' without a matching '}', making it invalid. This '{' is character %d. ",
			format, starts[end] + 1)
	}

	return starts, ends, nil
}

// splitFormatStrring splits a format string with variables at
// format[starts[i]: ends[i]]. The split fixed text and split variable text
// are returned in the original order. Variables do not have enclosing {...}
// around them. len(text) = len(vars) + 1. No errors are returned by this
// function because it is assumed that starts and ends were generated by
// startsEndsStringFormat
func splitFormatString(
	format string, starts, ends []int,
) (text, vars []string) {
	if len(starts) == 0 {
		return []string{format}, []string{ }
	}
	
	text = append(text, format[:starts[0]])

	for i := 0; i < len(starts) - 1; i++ {
		vars = append(vars, format[starts[i]+1: ends[i]-1])
		text = append(text, format[ends[i]: starts[i+1]])
	}

	// Handle the last var and text separately.
	end := len(ends) - 1
	vars = append(vars, format[starts[end]+1: ends[end]-1])
	if ends[end] < len(format) {
		text = append(text, format[ends[end]:])
	} else {
		text = append(text, "")
	}

	return text, vars
}

// splitFormatStringVar splits the components of the variable v into a printf
// verb and a variable
func splitFormatStringVar(
	v string, m map[string]int,
) (verb, rule string, isSeq bool, err error) {
	tok := strings.Split(v, ",")
	if len(tok) == 1 {
		return "", "", false, fmt.Errorf("it does not have a comma.")
	} else if len(tok) > 2 {
		return "", "", false, fmt.Errorf("it has more than one comma.")
	}

	verb, rule = tok[0], tok[1]
	verb, err = fixVerb(verb)
	if err != nil { return "", "", false, err }

	isSeq = !strings.ContainsAny(rule, AllLetters)

	if isSeq {		
		if _, err := ExpandSequenceFormat(rule); err != nil {
			return "", "", false, fmt.Errorf(
				"'%s' can't be parsed as a sequence. %s", rule, err.Error(),
			)
		}
	} else {
		if _, ok := m[rule]; !ok {
			return "", "", false, fmt.Errorf("'%s' is not a valid variable name. The only valid variable names are %s.",
				rule, allVars(m))
		}
	}

	return verb, rule, isSeq, nil
}

// allVars returns a sorted array of a map's keys.
func allVars(m map[string]int) []string {
	keys := []string{ }
	for key, _ := range m { keys = append(keys, key) }
	sort.Strings(keys)
	return keys
}

// fixVerb fixes printf verbs formatted after Python and C rules and throws an
// error if it's not a recognized command. This gets, ah, a little complicated.
func fixVerb(v string) (string, error) {
	err := fmt.Errorf("'%s' is not a valid printf() command.", v)
	
	if len(v) == 0 {
		return "", err
	} else if v[0] != '%' {
		return "", err
	}

	// In C, i and d are synonyms.
	v = strings.ReplaceAll(v, "i", "d") 
	// In C, l is required for long ints
	v = strings.ReplaceAll(v, "l", "") 
	// In C, h is required for short ints
	v = strings.ReplaceAll(v, "h", "")
	// In C, both . and 0 allow for zero-padding
	v = strings.ReplaceAll(v, ".", "0")
	// In Python there are obscure characters that no-one uses. Throw an error
	// if someone tries. In principle these aren't hard to implement, but I bet
	// no one will ever need them.
	if strings.ContainsAny(v, "<>=^") {
		return "", fmt.Errorf("'%s' contains obscure Python-2 format characters that Guppy hasn't implemented. Please submit an issue requesting this feature.", v)
	} else if strings.ContainsAny(v, "_,") {
		return "", fmt.Errorf("'%s' contains obscure Python-3 format characters that Guppy hasn't implemented. Please submit an issue requesting this feature.", v)
	}
		
	// Only ints are supported for now.
	switch v[len(v) - 1] {
	case 'b', 'c', 'd', 'o', 'O', 'q', 'x', 'X', 'U':
	default: return "", err
	}

	// Try our best to figure out if the modifier is legal.
	mod := v[1:len(v) - 1]
	prevFlag := "none"
ModLoop:
	for i, c := range mod {
		switch c {
		case '+', '-':
			if prevFlag != "none" {
				return "", err
			}
			prevFlag = "sign"
		case '#':
			if prevFlag != "none" && prevFlag != "sign" {
				return "", err
			}
			prevFlag = "alt"
		case '0', ' ':
			if prevFlag != "none" && prevFlag != "sign" && prevFlag != "alt" {
				return "", err
			}
			prevFlag = "padding"
		default:
			_, atoiErr:= strconv.Atoi(mod[i:len(mod)])
			if atoiErr != nil {
				return "", err
			}
			prevFlag = "paddingSize"

			break ModLoop
		}
	}
	
	return v, nil
}

// expandSeqValues multiplicatively expands seqeunce rules. For example, if you
// have sequence rules corresponding to [1, 2, 3] [4] [10, 1], you would get
// 1 4 10
// 2 4 10
// 3 4 10
// 1 4 11
// 2 4 11
// 3 4 11
// For convenience with other funcitons, these rules can be part of a rule
// array that does nto contain sequence rules, as specified by isSeq.
func expandSeqValues(rule []string, isSeq []bool) ([][]int, int) {
	if len(rule) == 0 {
		return [][]int{ }, 0
	}
	
	baseVals := make([][]int, len(isSeq))
	nTot := 1
	for i := range rule {
		if !isSeq[i] { continue }
		
		// The error value has already been checked.
		baseVals[i], _ = ExpandSequenceFormat(rule[i])
		nTot *= len(baseVals[i])
	}

	expVals := make([][]int, len(baseVals))
	for i := range baseVals {
		if !isSeq[i] { continue }

		nCurr := len(baseVals[i])
		nBase := 1
		for j := 0; j < i; j++ {
			if len(baseVals[j]) == 0 { continue }
			nBase *= len(baseVals[j])
		}

		expVals[i] = make([]int, nTot)
		for j := 0; j < nTot; j++ {
			expVals[i][j] = baseVals[i][(j / nBase) % nCurr]
		}
	}

	return expVals, nTot
}
