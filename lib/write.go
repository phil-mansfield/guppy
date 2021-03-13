package lib

import (
	read "github.com/phil-mansfield/guppy/go"
	"github.com/phil-mansfield/guppy/lib/snapio"
)

func CreateHeader(args *Args, origHd snapio.Header, i int) *read.Header {
	panic("NYI")
}

func Write(
	args *Args, file string, hd *read.Header,
	origHd snapio.Header, data []byte,
) {
	panic("NYI")
}
