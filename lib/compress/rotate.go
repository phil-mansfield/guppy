package compress

import (
	"fmt"
	"math"
)

const maxDelta = math.MaxInt16

type DeltaStats struct {
	hist, cSum [maxDelta + 1]int
	n int
}

func (stats *DeltaStats) Load(delta []int64) {
	stats.n = 0

	// Find how long the histogram is.
	for i := range delta {
		if delta[i] < 0 || delta[i] > maxDelta {
			panic(fmt.Sprint("Internal error: delta = %d is out of range for" +
				"the Rotator class. Guppy should have caught this earlier, " + 
				"so this is a bug, but the cause is probabaly that the user " +
				"set delta to be way too small.", delta[i]))
		}

		if int(delta[i] + 1) > stats.n { stats.n = int(delta[i] + 1) }
	}

	// Clear histogram
	for i := 0; i < stats.n; i++ { stats.hist[i] = 0 }

	// Update histogram
	for i := range delta { stats.hist[delta[i]]++ }

	// Update cumulative sum
	stats.cSum[0] = stats.hist[0]
	for i := 1; i < stats.n; i++ {
		stats.cSum[i] = stats.cSum[i-1] + stats.hist[i]
	}
}

func (stats *DeltaStats) Mean() int64 {
	sum := int64(0)
	n := int64(0)
	for i := 0; i < stats.n; i++ {
		sum += int64(stats.hist[i]*i)
		n += int64(stats.hist[i])
	}

	return sum / n
}

func (stats *DeltaStats) Mode() int64 {
	maxI := 0
	for i := 1; i < stats.n; i++ {
		if stats.hist[i] > stats.hist[maxI] { maxI = i }
	}
	return int64(maxI)
}

func (stats *DeltaStats) Window(size int) int64 {
	if size >= stats.n { return int64(stats.n / 2) }

	max := stats.cSum[size - 1]
	maxFirst := 0
	for first := 1; first + size - 1 < stats.n; first++ {
		diff := stats.cSum[first + size - 1] - stats.cSum[first - 1]
		if diff > max {
			maxFirst = first
			max = diff
		}
	}

	return int64(2*maxFirst + size) / 2
}

func MaxBytes(delta []int64) uint8 {
	max := delta[0]
	for i := 1; i < len(delta); i++ {
		if delta[i] > max { max =  delta[i] }
	}
	switch {
	case max <= math.MaxUint8: return 1 
	case max <= math.MaxUint16: return 2 
	case max <= math.MaxUint32: return 4 
	default: return 8
	}
}

func RotateEncode(maxBytes uint8, delta []int64, rot int64) {
	switch maxBytes {
	case 1:
		for i := range delta {
			delta[i] = int64(uint8(delta[i] - rot + 128))
		}
	case 2:
		for i := range delta {
			delta[i] = int64(uint16(delta[i] - rot + 128))
		}
	case 4:
		for i := range delta {
			delta[i] = int64(uint32(delta[i] - rot + 128))
		}
	default:
		for i := range delta {
			delta[i] = int64(uint64(delta[i] - rot + 128))
		}
	}
}

func RotateDecode(maxBytes uint8, delta []int64, rot int64) {
	switch maxBytes {
	case 1:
		for i := range delta {
			delta[i] = int64(uint8(delta[i]) - 128 + uint8(rot))
		}
	case 2:
		for i := range delta {
			delta[i] = int64(uint16(delta[i]) - 128 + uint16(rot))
		}
	case 4:
		for i := range delta {
			delta[i] = int64(uint32(delta[i]) - 128 + uint32(rot))
		}
	default:

		for i := range delta {
			delta[i] = delta[i] - 128 + rot
		}
	}
}