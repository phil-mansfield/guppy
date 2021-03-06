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

func TestSplitFormatString(t *testing.T) {
	tests := []struct {
		format string
		starts, ends []int
		text, vars []string
	} {
		{"", []int{}, []int{}, []string{""}, []string{}},
		{"aaaaa", []int{}, []int{},
			[]string{"aaaaa"}, []string{}},
		{"a{b}a", []int{1}, []int{4},
			[]string{"a", "a"}, []string{"b"}},
		{"{b}aa", []int{0}, []int{3},
			[]string{"", "aa"}, []string{"b"}},
		{"aa{b}", []int{2}, []int{5},
			[]string{"aa", ""}, []string{"b"}},
		{"{b}", []int{0}, []int{3},
			[]string{"", ""}, []string{"b"}},
		{"a{b}c(dd)e::", []int{1, 5, 10}, []int{4, 9, 12},
			[]string{"a", "c", "e", ""}, []string{"b", "dd", ""}},
	}

	for i := range tests {
		text, vars := splitFormatString(
			tests[i].format, tests[i].starts, tests[i].ends,
		)
		if !stringsEq(text, tests[i].text) || !stringsEq(vars, tests[i].vars) {
			t.Errorf("%d) Expected '%s', starts = %d, ends = %d to split to text = %s, vars = %s, but got text = %s, vars = %s",
				i, tests[i].format, tests[i].starts, tests[i].ends,
				tests[i].text, tests[i].vars, text, vars,
			)
		}
	}
}

func TestFixVerb(t *testing.T) {
	tests := []struct {
		in, out string
		valid bool
	} {
		{"%d", "%d", true},
		{"%c", "%c", true},
		{"%b", "%b", true},
		{"%o", "%o", true},
		{"%O", "%O", true},
		{"%q", "%q", true},
		{"%x", "%x", true},
		{"%X", "%X", true},
		{"%U", "%U", true},

		{"%i", "%d", true},
		{"%ld", "%d", true},
		{"%li", "%d", true},
		{"%hd", "%d", true},
		{"%hi", "%d", true},

		{"%+d", "%+d", true},
		{"%0d", "%0d", true},
		{"%03d", "%03d", true},
		{"%0100d", "%0100d", true},
		{"% 150d", "% 150d", true},
		{"%+031d", "%+031d", true},
		{"%-031d", "%-031d", true},
		{"%#031d", "%#031d", true},
		{"%+#031d", "%+#031d", true},
		{"%+#d", "%+#d", true},
		{"%#d", "%#d", true},
		{"%100d", "%100d", true},
		
		{"%+.41hi", "%+041d", true},
		{"%+#031li", "%+#031d", true},

		// I'm not going to be able to enumerate every incorrect format string
		{"", "", false},
		{"d", "", false},
		{"i", "", false},
		{"%f", "", false},
		{"%???d", "", false},
		{"%<d", "", false},
		{"%>d", "", false},
		{"%=d", "", false},
		{"%^d", "", false},
		{"%_d", "", false},
		{"%,d", "", false},
		{"%03+d", "", false},
		{"%0#d", "", false},
		{"%#+d", "", false},
	}

	for i := range tests {
		out, err := fixVerb(tests[i].in)
		if err != nil && tests[i].valid {
			t.Errorf("%d) Expected '%s' could be processed, but got error %s",
				i, tests[i].in, err.Error())
		} else if err == nil && !tests[i].valid {
			t.Errorf("%d) Expected '%s' would fail, but got no error.",
				i, tests[i].in)
		} else if out != tests[i].out {
			t.Errorf("%d) Expected '%s' would become '%s', but got '%s'.",
				i, tests[i].in, tests[i].out, out)
		}
	}
}

func TestSplitFormatStringVar(t *testing.T) {
	m :=  map[string]int{ "aa": 10, "bb": 20}
	tests := []struct{
		v, verb, rule string
		isSeq, valid bool
	} {
		{"%d,aa", "%d", "aa", false, true},
		{"%d,bb", "%d", "bb", false, true},
		{"%03d,aa", "%03d", "aa", false, true},
		{"%+li,aa", "%+d", "aa", false, true},
		{"%d,1", "%d", "1", true, true},
		{"%d,1..4", "%d", "1..4", true, true},
		{"% 10d,1..100 + 200..210 - 50", "% 10d", "1..100 + 200..210 - 50",
			true, true},

		{"%p,aa", "", "", false, false},
		{"%03+d,aa", "", "", false, false},
		{"%dd,aa", "", "", false, false},
		{"%p,cc", "", "", false, false},
		{"%p,1a", "", "", false, false},
		{"% 10d,1..100 + 200..210..1 - 50", "", "", false, false},

	}

	for i := range tests {
		verb, rule, isSeq, err := splitFormatStringVar(tests[i].v, m)

		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected '%s' could be split, but got error '%s'.",
				i, tests[i].v, err.Error())
		} else if !tests[i].valid && err ==nil {
			t.Errorf("%d) Expected '%s' couldn't be split, but go not error.",
				i, tests[i].v)
		} else if tests[i].valid && (isSeq != tests[i].isSeq ||
			verb != tests[i].verb || rule != tests[i].rule) {
			t.Errorf("%d) Expected '%s' would be split into verb = '%s', rule = '%s', and isSeq = %v, but got verb = '%s', rule = '%s', and isSeq = %v.",
				i, tests[i].v, tests[i].verb, tests[i].rule,
				tests[i].isSeq, verb, rule, isSeq)
		}
	}
}

func TestExpandSeqeunceValues(t *testing.T) {
	tests := []struct {
		rule []string
		isSeq []bool
		expSeq [][]int
		nTot int
	} {
		{
			[]string{},
			[]bool{},
			[][]int{},
			0,
		},
		{
			[]string{"1"},
			[]bool{true},
			[][]int{{1}},
			1,
		},
		{
			[]string{"1 + 3..5"},
			[]bool{true},
			[][]int{{1, 3, 4, 5}},
			4,
		},
		{
			[]string{"1", "2"},
			[]bool{true, false},
			[][]int{{1}, {}},
			1,
		},
		{
			[]string{"1", "2"},
			[]bool{false, true},
			[][]int{{}, {2}},
			1,
		},
		{
			[]string{"1", "2"},
			[]bool{true, true},
			[][]int{{1}, {2}},
			1,
		},
		{
			[]string{"1", "2", "3"},
			[]bool{true, false, true},
			[][]int{{1}, {}, {3}},
			1,
		},
		{
			[]string{"1..3", "10..11",},
			[]bool{true, true},
			[][]int{{1, 2, 3, 1, 2, 3}, {10, 10, 10, 11, 11, 11}},
			6,
		},
		{
			[]string{"1..3", "10..11", "0 + 5"},
			[]bool{true, true, false},
			[][]int{{1, 2, 3, 1, 2, 3}, {10, 10, 10, 11, 11, 11}, {}},
			6,
		},
		{
			[]string{"1..3", "4", "10..11", "0 + 5"},
			[]bool{true, true, true, true},
			[][]int{{1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3},
				{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
				{10, 10, 10, 11, 11, 11, 10, 10, 10, 11, 11, 11},
				{0, 0, 0, 0, 0, 0, 5, 5, 5, 5, 5, 5}},
			12,
		},
	}

	for i := range tests {
		expSeq, nTot := expandSeqValues(tests[i].rule, tests[i].isSeq)

		if nTot != tests[i].nTot {
			t.Errorf("%d) Expected %s to expand to %d elemnts, got %d.",
				i, tests[i].rule, tests[i].nTot, nTot)
		} else if !intArraysEq(expSeq, tests[i].expSeq) {
			t.Errorf("%d) Expected %s to expand to %d, got %d.",
				i, tests[i].rule, tests[i].expSeq, expSeq)
		}
	}
}

func TestExpandFormatString(t *testing.T) {
	m := map[string]int{ "aa": 13, "bb": 42 }
	tests := []struct {
		format string
		out []string
		valid bool
	} {
		{"", []string{""}, true},
		{"qwerty", []string{"qwerty"}, true},
		{"{%d,aa}", []string{"13"}, true},
		{"{%d,bb}", []string{"42"}, true},
		{"{%03d,aa}", []string{"013"}, true},
		{"{%03li,aa}", []string{"013"}, true},
		{"{%d,aa}:{%d,bb}", []string{"13:42"}, true},
		{"{%d,aa} {%d,bb}", []string{"13 42"}, true},
		{"{%03d,aa} {%03d,bb}: {%+d,aa}", []string{"013 042: +13"}, true},
		{"{%d,1}", []string{"1"}, true},
		{"{%03d,20}", []string{"020"}, true},
		{"?{%d,25}{% 5d,bb}?", []string{"?25   42?"}, true},
		{"petting {%d,3} cats and {%d,2} dogs",
			[]string{"petting 3 cats and 2 dogs"}, true},
		{"{%d,1..3}", []string{"1", "2", "3"}, true},
		{"({%d,aa}^{%d,1..3})", []string{"(13^1)", "(13^2)", "(13^3)"}, true},
		{"{%d,1..2}{%d,3..4}{%d,5..6}",
			[]string{"135", "235", "145", "245",
				"136", "236", "146", "246"}, true},
		{"{%d,1..2} {%03d,aa} {%d,5..6}",
			[]string{"1 013 5", "2 013 5", "1 013 6", "2 013 6"}, true},

		{"wombo{combo", nil, false},
		{"wombo}combo", nil, false},
		{"wombo{}combo", nil, false},
		{"{wombo}combo", nil, false},
		{"{wombo,combo}", nil, false},
		{"{%wombo,combo}", nil, false},
		{"wombo {%d,combo}", nil, false},
		{"wombo {%combo,1}", nil, false},
		{"wombo {%^d,1} combo", nil, false},
		{"wombo {%0?d,1} combo", nil, false},
		{"wombo {%d,1..2..3} combo", nil, false},
		{"wombo {%d,1..2 - 3} combo", nil, false},
	}

	for i := range tests {
		out, err := ExpandFormatString(tests[i].format, m)
		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected '%s' could be expanded, but got error %s.",
				i, tests[i].format, err.Error())
		} else if !tests[i].valid && err == nil {
			t.Errorf("%d) Expected '%s' would fail, but got not error.",
				i, tests[i].format)
		} else if !stringsEq(out, tests[i].out) {
			t.Errorf("%d) Expected '%s' would expand to %s, but got %s.",
				i, tests[i].format, tests[i].out, out)
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

func intArraysEq(x, y [][]int) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if !intsEq(x[i], y[i]) { return false }
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
