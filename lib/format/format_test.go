package format

import (
	"testing"
)

func TestIsSequenceFormatToken(t *testing.T) {
	tests := []struct{
		tok string
		valid bool
	} {
		{"", false},
		{"1", true},
		{"a", false},
		{"1..30", true},
		{"a..30", false},
		{"1..a", false},
		{"30..1", false},
		{"a..b", false},
		{"1..30..60", false},
	}

	for i := range tests {
		err := isSequenceFormatToken(tests[i].tok)
		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected token '%s' to be valid, but got error '%s'.",
				i, tests[i].tok, err.Error())
		} else if !tests[i].valid && err == nil {
			t.Errorf("%d) Expected token '%s' to be invalid, but got no error.",
				i, tests[i].tok)
		}
	}
}

func TestParseSeqeunceFormatToken(t *testing.T) {
	tests := []struct{
		tok string
		seq []int
	} {
		{"0", []int{0}},
		{"1000", []int{1000}},
		{"1..4", []int{1, 2, 3, 4}},
		{"10..20", []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
	}

	for i := range tests {
		seq := parseSequenceFormatToken(tests[i].tok)
		if !intsEq(tests[i].seq, seq) {
			t.Errorf("%d) Expected token '%s' to expand to %d, got %d.",
				i, tests[i].tok, tests[i].seq, seq)
		}
	}
}

func TestTokeniseSequenceFormat(t *testing.T) {
	tests := []struct{
		format string
		tok []string
		valid bool
	} {
		{"", []string{""}, false},
		{"0", []string{"0"}, true},
		{"101", []string{"101"}, true},
		{"10..20", []string{"10..20"}, true},
		{"a..b", []string{"a..b"}, true},
		{"0+1", []string{"0", "+", "1"}, true},
		{"0 + 1", []string{"0", "+", "1"}, true},
		{"0-1", []string{"0", "-", "1"}, true},
		{"0 - 1", []string{"0", "-", "1"}, true},
		{"  0+       1    ", []string{"0", "+", "1"}, true},
		{"-0..100 + 0..200-9", []string{"-", "0..100", "+", "0..200",
			"-", "9"}, true},
		{"+-+-", []string{"+", "-", "+", "-"}, true},
	}

	for i := range tests {
		tok, err := tokeniseSequenceFormat(tests[i].format)
		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected '%s' to be valid, but got error '%s'.",
				i, tests[i].tok, err.Error())
		} else if !tests[i].valid && err == nil {
			t.Errorf("%d) Expected '%s' to be invalid, but got no error.",
				i, tests[i].tok)
		}

		if tests[i].valid && !stringsEq(tok, tests[i].tok) {
			t.Errorf("%d) Expected '%s' to tokenize to %s, got %s.",
				i, tests[i].format, tests[i].tok, tok)
		}
	}
}

func TestAddsSubsSequenceFormat(t *testing.T) {
	tests := []struct {
		tok, adds, subs []string
		valid bool
	} {
		{[]string{}, nil, nil, false},
		{[]string{"1"}, []string{"1"}, []string{}, true},
		{[]string{"+", "1"}, []string{"1"}, []string{}, true},
		{[]string{"-", "1"}, []string{}, []string{"1"}, true},
		{[]string{"1", "+", "2"}, []string{"1", "2"}, []string{}, true},
		{[]string{"1", "+", "2..10"}, []string{"1", "2..10"}, []string{}, true},
		{[]string{"1", "-", "2"}, []string{"1"}, []string{"2"}, true},
		{[]string{"1", "-", "2..10"}, []string{"1"}, []string{"2..10"}, true},
		{[]string{"-", "1", "-", "2"}, []string{}, []string{"1", "2"}, true},
		{[]string{"1", "2"}, nil, nil, false},
		{[]string{"1", "+", "2", "+"}, nil, nil, false},
		{[]string{"1", "+", "+", "2"}, nil, nil, false},
		{[]string{"1", "-", "-", "2"}, nil, nil, false},
		{[]string{"1", "+", "-", "2"}, nil, nil, false},
		{[]string{"1", "+"}, nil, nil, false},
		{[]string{"1", "-", "+", "2"}, nil, nil, false},
		{[]string{"1", "*", "2"}, nil, nil, false},
		{[]string{"+", "+", "1", "+", "2"}, nil, nil, false},
		{[]string{"a", "+", "2"}, nil, nil, false},
		{[]string{"1", "+", "a"}, nil, nil, false},
		{[]string{"1", "+", "a..2"}, nil, nil, false},
	}

	for i := range tests {
		adds, subs, err := addsSubsSequenceFormat(tests[i].tok)
		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected %s could be processed, got error '%s'",
				i, tests[i].tok, err.Error())
		} else if !tests[i].valid && err == nil{
			t.Errorf("%d) Expected %s could not be process, but got no error.",
				i, tests[i].tok)
		} else if tests[i].valid && (!stringsEq(adds, tests[i].adds) ||
			!stringsEq(subs, tests[i].subs)) {
			t.Errorf("%d) Expected %s would be processed into adds = %s, subs= %s, but got adds = %s, subs = %s",
				i, tests[i].tok, tests[i].adds, tests[i].subs, adds, subs)
		}
	}
}

func TestExpandSeqeunceFormat(t *testing.T) {
	tests := []struct{
		format string
		n []int
		valid bool
	} {
		{"", nil, false},
		{"a", nil, false},
		{"10..a", nil, false},
		{"a..10", nil, false},
		{"1", []int{ 1 }, true},
		{"1..5", []int{ 1, 2, 3, 4, 5 }, true},
		{"+1", []int{ 1 }, true},
		{"+1..5", []int{ 1, 2, 3, 4, 5 }, true},
		{"+ 1", []int{ 1 }, true},
		{"+ 1..5", []int{ 1, 2, 3, 4, 5 }, true},
		{"-1", nil, false},
		{"-1..5", nil, false},
		{"- 1", nil, false},
		{"- 1..5", nil, false},
		{"1 + 2", []int{1, 2}, true},
		{"1+2", []int{1, 2}, true},
		{"1 +2", []int{1, 2}, true},
		{"1+ 2", []int{1, 2}, true},
		{"1 + 1", []int{1}, false},
		{"1 + 3..5", []int{1, 3, 4, 5}, true},
		{"3..5 + 1", []int{1, 3, 4, 5}, true},
		{"3..5 + 1 + 7..9", []int{1, 3, 4, 5, 7, 8, 9}, true},
		{"-3 + 3..5 - 4", []int{5}, true},
		{"1..10 - 2..9", []int{1, 10}, true},
		{"3..5 - 1", nil, false},
		{"3..5 - 4 - 4", nil, false},
		{"3..5 + 6+", nil, false},
		{"3..5 + 6-", nil, false},
	}

	for i := range tests {
		n, err := ExpandSequenceFormat(tests[i].format)

		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected '%s' could be expanded, got error '%s'",
				i, tests[i].format, err.Error())
		} else if !tests[i].valid && err == nil{
			t.Errorf("%d) Expected '%s' should fail, but got no error.",
				i, tests[i].format)
		} else if tests[i].valid && !intsEq(n, tests[i].n) {
			t.Errorf("%d) Expected '%s' to expand to %d, got %d",
				i, tests[i].format, tests[i].n, n)
		}
	}
}

func TestStartsEndsFormatString(t *testing.T) {
	tests := []struct{
		format string
		starts, ends []int
		valid bool
	} {
		{"aaaaaa", []int{}, []int{}, true},
		{"a{bb}a", []int{1}, []int{5}, true},
		{"{bb}aa", []int{0}, []int{4}, true},
		{"aa{bb}", []int{2}, []int{6}, true},
		{"{}", []int{0}, []int{2}, true},
		{"{}{bb}{}{}", []int{0, 2, 6, 8}, []int{2, 6, 8, 10}, true},
		{"{}{bb}a{}{}", []int{0, 2, 7, 9}, []int{2, 6, 9, 11}, true},
		{"{", nil, nil, false},
		{"}", nil, nil, false},
		{"{{", nil, nil, false},
		{"{{}}", nil, nil, false},
		{"{}{", nil, nil, false},
		{"{}}", nil, nil, false},
		{"}{}", nil, nil, false},
		{"{{}", nil, nil, false},
	}

	for i := range tests {
		starts, ends, err := startsEndsFormatString(tests[i].format)
		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected '%s' could be processed, but got error '%s'",
				i, tests[i].format, err.Error())
		} else if !tests[i].valid && err == nil {
			t.Errorf("%d) Expected '%s' should fail, but got no error.",
				i, tests[i].format)
		} else if !intsEq(starts, tests[i].starts) ||
			!intsEq(ends, tests[i].ends) {
			t.Errorf("%d) Expected '%s' should have starts = %d, ends = %d, but got starts = %d, ends = %d",
				i, tests[i].format, tests[i].starts,
				tests[i].ends, starts, ends,
			)
		}
	}
}

//////////////////////
// Helper functions //
//////////////////////

func intsEq(x, y []int) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func stringsEq(x, y []string) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
