package read_guppy

import (
	"math"
)

const (
	BoundsBins = 128
)

// periodicRangeContains returns true if x is within [start, end) and false
// otherwise.
func periodicRangeContains(start, end, x, L float32)  bool {
	if end < L {
		return start <= x && end > x
	} else {
		return end - L > x || start <= x
	}
}

// periodiRangeOverlap returns true if two periodic ranges [start1, end1)
// and [start2, end2) overlap and false otherwise.
func periodicRangeOverlap(start1, end1, start2, end2, L float32) bool {
	return periodicRangeContains(start1, end1, start2, L) ||
		periodicRangeContains(start2, end2, start1, L) 
		
}

// PeriodicOverlap returns true if two 3D periodic bounding boxes
// overlap and false otherwise. The bounds are formatted as [start_x, start_y,
// start_z, end_x, end_y, end_z). start_i must be in [0, L) and end_i must be in
// the range (start_i, start_i + L). b2 is expanded by buf in each dimension
// and is written to b3.
func PeriodicOverlap(b1, b2 []float32, L float32) bool {
	overlap := true
	for dim := 0; dim < 3; dim++ {
		overlap = overlap && periodicRangeOverlap(b1[dim], b1[dim+3],
			b2[dim], b2[dim+3], L)
	}
	return overlap
}

// binRanges computes the maximum and minimum value in each bin and assigns
// those values to maxes and mins, respectively. If there are no particles
// in a bin, mins is set to -1 and maxes is set to L.
func binRanges(x [][3]float32, L float32, dim int, mins, maxes []float32) {
	// Sentinel values indicating that there are no particles in these bins.
	for i := range mins {
		mins[i] = -1
		maxes[i] = -1
	}
	
	bins := len(mins)
	dx := L / float32(bins)
	
	for i := range x {
		xi := x[i][dim];

		var binIdx int
		if xi < 0 {
			binIdx = int(math.Floor(float64(xi / dx)))
		} else {
			binIdx = int(xi / dx)
		}
		
		if binIdx < 0 { binIdx += bins }
		if binIdx >= bins { binIdx -= bins }
		
		if mins[binIdx] == -1 || xi < mins[binIdx] { mins[binIdx] = xi }
		if mins[binIdx] == -1 || xi > maxes[binIdx] { maxes[binIdx] = xi }
	}
}

// startingBin returns the index of the first bin with particles in it.
func startingBin(mins []float32) int {
	for i := range mins {
		if mins[i] >= 0 { return i }
	}
	return -1
}

// endingBin returns the index of the last bin with particles in it.
func endingBin(mins []float32) int {
	for i := len(mins) - 1; i >= 0; i-- {
		if (mins[i] >= 0) { return i }
	}
	return -1
}

// nextUsedBin returns the next bin after i which has particles in it.
func nextUsedBin(mins []float32, i int) int {
	for j := i + 1; j < len(mins); j++ {
		if mins[j] >= 0 { return j }
	}
	return -1
}

// periodicRange is a helper-function for PeriodicBounds, which computes
// the 1D bounding range for x along the given dimension, dim. Otherwise, it
// acts identically.
func periodicRange(x [][3]float32, L float32, dim int, bounds []float32) {
	mins, maxes := make([]float32, BoundsBins), make([]float32, BoundsBins)
  
	binRanges(x, L, dim, mins, maxes)
	start := startingBin(mins)
	end := endingBin(mins)
	
	maxGapWidth := L + mins[start] - maxes[end]
	maxGapStart := maxes[end]

	// i points to the start of each gap
	for i := start; i < end; {
		// j points to the end of each gap
		j := nextUsedBin(mins, i)
		if j == -1 { break }
		
		if mins[j] - maxes[i] > maxGapWidth {
			maxGapStart = maxes[i]
			maxGapWidth = mins[j] - maxes[i]
		}
		
		i = j
	}
	
	bounds[dim] = maxGapStart + maxGapWidth
	if bounds[dim] > L { bounds[dim] -= L }
	bounds[dim + 3] = bounds[dim] + L - maxGapWidth
}

// PeriodicBounds returns the smallest bounding box that encloses all the
// points in x within a periodic box with dimensions [0, L)^3. The bounding box
// is output as a length-6 array, The first three indexes given the "lower"
// corner of the box, which will always be in [0, L)^3, and the upper three 
// give the "upper" corner, which may contain values greater than L if the 
// particle distribution spans the boundary of the box.
func PeriodicBounds(x [][3]float32, L float32) []float32 {
	bounds := make([]float32, 6)
	for dim := 0; dim < 3; dim++ {
		periodicRange(x, L, dim, bounds)
	}

	return bounds
}
