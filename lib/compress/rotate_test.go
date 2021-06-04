package compress

import (
	"math"
	"testing"
	"github.com/phil-mansfield/guppy/lib/eq"
)

func TestMaxBytes(t *testing.T) {
	tests := []struct{
		n int64
		bytes uint8
	} {
		{0, 1},
		{math.MaxInt8, 1},
		{math.MaxUint8, 1},
		{math.MaxUint8+1, 2},
		{math.MaxInt16, 2},
		{math.MaxUint16, 2},
		{math.MaxUint16+1, 4},
		{math.MaxInt32, 4},
		{math.MaxUint32, 4},
		{math.MaxUint32+1, 8},
		{math.MaxInt64, 8},
	}

	for i := range tests {

		delta := []int64{1, 0, 1, tests[i].n}

		if MaxBytes(delta) != tests[i].bytes {
			t.Errorf("%d.1) Expected MaxBytes(%d) = %d, got %d.", i, 
				tests[i].n, tests[i].bytes, MaxBytes(delta))
		}

		delta = []int64{tests[i].n, 1, 0, 1}

		if MaxBytes(delta) != tests[i].bytes {
			t.Errorf("%d.2) Expected MaxBytes(%d) = %d, got %d.", i, 
				tests[i].n, tests[i].bytes, MaxBytes(delta))
		}
	}
}

func TestLoad(t *testing.T) {
	tests := []struct{
		delta []int64
		hist []int
	} {
		{[]int64{}, []int{}},
		{[]int64{0}, []int{1}},
		{[]int64{0, 0}, []int{2}},
		{[]int64{0}, []int{1}},
		{[]int64{0, 1, 0}, []int{2, 1}},
		{[]int64{0, 1, 0}, []int{2, 1}},
		{[]int64{0, 2, 4, 2, 2}, []int{1, 0, 3, 0, 1}},
		{[]int64{5}, []int{0, 0, 0, 0, 0, 1}},
	}

	stats := &DeltaStats{ }
	for i := range tests {
		stats.Load(tests[i].delta)
		hist := stats.hist[:stats.n]

		if !eq.Ints(hist, tests[i].hist) {
			t.Errorf("%d) Expected delta = %d to give the histogram %d. " + 
				"Got %d instead.", i, tests[i].delta, tests[i].hist, hist)
		}
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		delta []int64
		mean int64
	} {
		{[]int64{0}, 0},
		{[]int64{300}, 300},
		{[]int64{5, 11}, 8},
		{[]int64{5, 10}, 7},
		{[]int64{0, 1, 1, 1, 1, 3, 3, 3, 3, 4}, 2},
	}

	stats := &DeltaStats{ }
	for i := range tests {
		stats.Load(tests[i].delta)
		mean := stats.Mean()

		if tests[i].mean != mean {
			t.Errorf("%d) Expected mean of %d to be %d, got %d.", i,
				tests[i].delta, tests[i].mean, mean)
		}
	}
}

func TestMode(t *testing.T) {
	tests := []struct {
		delta []int64
		mode int64
	} {
		{[]int64{0}, 0},
		{[]int64{300}, 300},
		{[]int64{5, 11}, 5},
		{[]int64{5, 10, 10, 7}, 10},
		{[]int64{0, 1, 1, 1, 1, 3, 3, 3, 3, 3, 4, 1, 1}, 1},
	}

	stats := &DeltaStats{ }
	for i := range tests {
		stats.Load(tests[i].delta)
		mode := stats.Mode()

		if tests[i].mode != mode {
			t.Errorf("%d) Expected mode of %d to be %d, got %d.", i,
				tests[i].delta, tests[i].mode, mode)
		}
	}
}

func TestWindow(t *testing.T) {
	tests := []struct{
		hist []int
		size int
		center int64
	} {
		{[]int{1, 1}, 10, 1},
		{[]int{1, 1, 1}, 10, 1},
		{[]int{1, 1, 1, 1}, 10, 2},
		{[]int{2, 2, 2, 1, 1}, 3, 1},
		{[]int{1, 2, 2, 2, 1}, 3, 2},
		{[]int{1, 1, 2, 2, 2}, 3, 3},
		{[]int{2, 2, 1, 1, 1}, 2, 1},
		{[]int{1, 1, 2, 2, 1}, 2, 3},
		{[]int{1, 1, 1, 2, 2}, 2, 4},
	}

	stats := &DeltaStats{ }

	for i := range tests {
		delta := []int64{ }
		for j := range tests[i].hist {
			for k := 0; k < tests[i].hist[j]; k++ {
				delta = append(delta, int64(j))
			}
		}

		stats.Load(delta)
		center := stats.Window(tests[i].size)

		if center != tests[i].center {
			t.Errorf("%d) Expected center = %d, got %d.",
				i, tests[i].center, center)
		} 
	}
}

func TestRotateEncode(t *testing.T) {
	tests := []struct{
		delta []int64
		rot int64
		rotDelta []int64
	} {
		{[]int64{0xfe, 0xff, 0, 1, 2, 3}, 128, []int64{0xfe, 0xff, 0, 1, 2, 3}},
		{[]int64{0xfe, 0xff, 0, 1, 2, 3}, 121, []int64{5, 6, 7, 8, 9, 10}},
		{[]int64{0xfe, 0xff, 0, 1, 2, 3}, 130,
			[]int64{0xfc, 0xfd, 0xfe, 0xff, 0, 1}},
		{[]int64{0xfffe, 0xffff, 0, 1, 2}, 128,
			[]int64{0xfffe, 0xffff, 0, 1, 2}},
		{[]int64{0xfffe, 0xffff, 0, 1, 2}, 127,
			[]int64{0xffff, 0, 1, 2, 3}},
		{[]int64{0xfffe, 0xffff, 0, 1, 2}, 129,
			[]int64{0xfffd, 0xfffe, 0xffff, 0, 1}},
	}

	for i := range tests {
		rotDelta := make([]int64, len(tests[i].delta))
		copy(rotDelta, tests[i].delta)

		maxBytes := MaxBytes(rotDelta)
		RotateEncode(maxBytes, rotDelta, tests[i].rot)

		if !eq.Int64s(rotDelta, tests[i].rotDelta) {
			t.Errorf("%d) Expected delta = %04x, rot = %d -> " + 
				"rotDelta = %04x, got %d", i, tests[i].delta, tests[i].rot,
				tests[i].rotDelta, rotDelta)
		}
	}
}

func TestRotateDecode(t *testing.T) {
	tests := []struct{
		delta []int64
		rot int64
		rotDelta []int64
		maxBytes uint8
	} {
		{[]int64{0xfe, 0xff, 0, 1, 2, 3}, 128,
			[]int64{0xfe, 0xff, 0, 1, 2, 3}, 1},
		{[]int64{0xfe, 0xff, 0, 1, 2, 3}, 121,
			[]int64{5, 6, 7, 8, 9, 10}, 1},
		{[]int64{0xfe, 0xff, 0, 1, 2, 3}, 130,
			[]int64{0xfc, 0xfd, 0xfe, 0xff, 0, 1}, 1},
		{[]int64{0xfffe, 0xffff, 0, 1, 2}, 128,
			[]int64{0xfffe, 0xffff, 0, 1, 2}, 2},
		{[]int64{0xfffe, 0xffff, 0, 1, 2}, 127,
			[]int64{0xffff, 0, 1, 2, 3}, 2},
		{[]int64{0xfffe, 0xffff, 0, 1, 2}, 129,
			[]int64{0xfffd, 0xfffe, 0xffff, 0, 1}, 2},
	}

	for i := range tests {
		delta := make([]int64, len(tests[i].delta))
		copy(delta, tests[i].rotDelta)

		RotateDecode(tests[i].maxBytes, delta, tests[i].rot)

		if !eq.Int64s(delta, tests[i].delta) {
			t.Errorf("%d) Expected rotDelta = %04x, rot = %d -> " + 
				"rotDelta = %04x, got %d", i, tests[i].rotDelta, tests[i].rot,
				tests[i].delta, delta)
		}
	}
}