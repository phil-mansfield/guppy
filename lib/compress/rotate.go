package compress

// DeltaStats is a histogram containing delta values which can be used
// to compute various statistics of the delta distribution.
type DeltaStats struct {
	hist, cSum []int
	nMin, nMax int64
}

// Load loads an array into the DeltaStas array. It must be called
// before other methods are called.
func (stats *DeltaStats) Load(delta []int64) {
	if len(delta) == 0 {
		stats.nMin, stats.nMax = 0, 0
		stats.hist, stats.cSum = stats.hist[:0], stats.cSum[:0]
		return
	}

	stats.nMin, stats.nMax = delta[0], delta[0]

	// Find how long the histogram is.
	for i := range delta {
		if delta[i] < stats.nMin {
			stats.nMin = delta[i]
		} else if delta[i] > stats.nMax {
			stats.nMax = delta[i]
		}
	}

	n := stats.nMax - stats.nMin + 1
	stats.hist = expandInts(stats.hist, int(n))
	stats.cSum = expandInts(stats.cSum, int(n))

	// Clear histogram
	for i := range stats.hist { stats.hist[i] = 0 }

	// Update histogram
	for i := range delta { stats.hist[delta[i] - stats.nMin]++ }

	// Update cumulative sum
	stats.cSum[0] = stats.hist[0]
	for i := 1; i < len(stats.cSum); i++ {
		stats.cSum[i] = stats.cSum[i-1] + stats.hist[i]
	}
}

// expandInts expands x to have length n, making the minimum number of
// heap allocations.
func expandInts(x []int, n int) []int {
	if x == nil { return make([]int, n) }

	if cap(x) >= n {
		x = x[:n]
	} else {
		x = x[:cap(x)]
		x = append(x, make([]int, n - cap(x))...)
	}

	return x
}

// Mean returns the mean of the histogram.
func (stats *DeltaStats) Mean() int64 {
	sum := int64(0)
	n := int64(0)
	for i := range stats.hist {
		sum += int64(stats.hist[i]*(i + int(stats.nMin)))
		n += int64(stats.hist[i])
	}

	return sum / n
}

// Mode returns the mode of the histogram.
func (stats *DeltaStats) Mode() int64 {
	maxI := 0
	for i := range stats.hist {
		if stats.hist[i] > stats.hist[maxI] { maxI = i }
	}
	return int64(maxI) + stats.nMin
}

// Window returns the center of "window" of the given size which contains
// the maximum number of values.
func (stats *DeltaStats) Window(size int) int64 {
	if size >= len(stats.hist) {
		return int64(len(stats.hist)) / 2 + stats.nMin
	}

	max := stats.cSum[size - 1]
	maxFirst := 0
	for first := 1; first + size - 1 < len(stats.hist); first++ {
		diff := stats.cSum[first + size - 1] - stats.cSum[first - 1]
		if diff > max {
			maxFirst = first
			max = diff
		}
	}

	return (2*(stats.nMin + int64(maxFirst)) + int64(size)) / 2
}

// NeededRotation returns how many values higher the each element in delta
// woudl need to shift to make sure that all elements are positive and
// that mid % 256 = 127. If mid is chosen approppriately, this latter
// condition can allow zlib to compress values more efficeintly.
func (stats *DeltaStats) NeededRotation(mid int64) int64 {
	// This would be the rotation if we didn't care about aligning 
	// mid to the middle of a byte.
	offset := -int64(stats.nMin)

	// Rotation needed to align mid to the middle of a byte. Sorry, but the
	// next couple lines are confusing.
	var centering int64

	midMod := (offset + mid) % 256
	if midMod < 0 { midMod += 256 }
	centering = 127 - midMod
	if centering < 0 { centering += 256 }

	return offset + centering
}

func RotateEncode(delta []int64, rot int64) {
	for i := range delta { delta[i] += rot }
}

func RotateDecode(delta []int64, rot int64) {
	for i := range delta { delta[i] -= rot }
}
