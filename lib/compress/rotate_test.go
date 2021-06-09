package compress

import (
	"testing"
	"github.com/phil-mansfield/guppy/lib/eq"
)

func TestLoad(t *testing.T) {
	tests := []struct{
		delta []int64
		hist []int
		nMin int
	} {
		{[]int64{}, []int{}, 0},
		{[]int64{0}, []int{1}, 0},
		{[]int64{0, 0}, []int{2}, 0},
		{[]int64{0}, []int{1}, 0},
		{[]int64{0, 1, 0}, []int{2, 1}, 0},
		{[]int64{0, 1, 0}, []int{2, 1}, 0},
		{[]int64{0, 2, 4, 2, 2}, []int{1, 0, 3, 0, 1}, 0},
		{[]int64{5}, []int{1}, 5},
		{[]int64{3, 4, 5, 5}, []int{1, 1, 2}, 3},
		{[]int64{-3, 0, -3}, []int{2, 0, 0, 1}, -3},
	}

	stats := &DeltaStats{ }
	for i := range tests {
		stats.Load(tests[i].delta)
		if !eq.Ints(stats.hist, tests[i].hist) {
			t.Errorf("%d) Expected delta = %d to give the histogram %d. " + 
				"Got %d instead.", i, tests[i].delta, tests[i].hist, stats.hist)
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

func TestNeededRotation(t *testing.T) {
	nMinLow, nMinHigh := -256, 256
	midHigh := int64(512)

	stats := &DeltaStats{ }

	for nMin := nMinLow; nMin <= nMinHigh; nMin++ {
		for mid := int64(nMin); mid <= midHigh; mid++ {
			stats.nMin = int64(nMin)
			rot := stats.NeededRotation(mid)

			if (rot + mid) % 256 != 127 {
				t.Fatalf("err 1: mid = %d, nMin = %d gave rot = %d",
					mid, nMin, rot)
			}
			if rot + int64(nMin) < 0 {
				t.Fatalf("err 2: mid = %d, nMin = %d gave rot = %d",
					mid, nMin, rot)	
			}
		}
	}
}