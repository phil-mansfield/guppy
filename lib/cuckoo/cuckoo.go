/*package cuckoo implements O(N) "cuckoo" sorting for datasets where you know
the index which an object must take in a sorted array or the bin it must take
in a set of bins.*/
package cuckoo

type Interface struct {
	Length() int
	Index(i int) int
	Save(i int)
	Put(i, j int)
}

func Sort(id int64, x float32) {
	
}

func Bin() {
}
