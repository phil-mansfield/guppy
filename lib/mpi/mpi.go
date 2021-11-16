package main

// This header is almost the same as the one used by
// github.com/marcusthierfelder/mpi with some minor changes as well as a
// changes to the way that compilation is done. I'd import this package like
// normal, but these changes impact the underlying type system and compilation
// instructions, so that's not possible. As such, here is his license:
//
// Copyright (c) 2017 Marcus Thierfelder
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// NOTE: Use
// $ mpicc --showme:compile
// $ mpicc --showme:link
// To figure out CFLAGS and LDFLAGS, respectively

/*
#cgo LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
#cgo CFLAGS: -std=gnu99 -Wall -I/usr/lib/x86_64-linux-gnu/openmpi/include/openmpi -I/usr/lib/x86_64-linux-gnu/openmpi/include -pthread
#include <mpi.h>
#include <stdlib.h>

MPI_Comm get_MPI_COMM_WORLD() {
    return (MPI_Comm)(MPI_COMM_WORLD);
}

MPI_Datatype get_MPI_Datatype(int i) {
    switch(i) {
    case 0: return (MPI_Datatype)MPI_INT;
    case 1: return (MPI_Datatype)MPI_LONG_LONG;
    case 2: return (MPI_Datatype)MPI_FLOAT;
    case 3: return (MPI_Datatype)MPI_DOUBLE;
    }
    return NULL;
}
*/
import "C"

// The variables here and the first four functions are taken verbatim from
// github.com/marcusthierfelder/mpi. Later functions are new additions (based
// on the function design in the original package).

import (
	"fmt"

	"unsafe"
)

var (
	COMM_WORLD C.MPI_Comm = C.get_MPI_COMM_WORLD()

	INT32   C.MPI_Datatype = C.get_MPI_Datatype(0)
	INT64   C.MPI_Datatype = C.get_MPI_Datatype(1)
	FLOAT32 C.MPI_Datatype = C.get_MPI_Datatype(2)
	FLOAT64 C.MPI_Datatype = C.get_MPI_Datatype(3)
)

// These functions have modified error reporting

func Comm_size(comm C.MPI_Comm) int {
	n := C.int(-1)
	err := C.MPI_Comm_size(comm, &n)
	processError(err)
	return int(n)
}

func Comm_rank(comm C.MPI_Comm) int {
	n := C.int(-1)
	err := C.MPI_Comm_rank(comm, &n)
	processError(err)
	return int(n)
}

func Init() {
	err := C.MPI_Init(nil, nil)
	processError(err)
}

func Finalize() {
	err := C.MPI_Finalize()
	processError(err)
}

// These functions are new.

func processError(err C.int) {
	if err == 0 { return }
	
	buf := make([]C.char, C.MPI_MAX_ERROR_STRING)
	n := C.int(0)
	C.MPI_Error_string(err, &buf[0], &n)
	panic(C.GoString(&buf[0]))
}

func Bcast_int64(buffer []int64, root int, comm C.MPI_Comm) {
	err := C.MPI_Bcast(unsafe.Pointer(&buffer[0]), C.int(len(buffer)),
		INT64, C.int(root), comm)
	processError(err)
}

func Gather_int64(send, recv []int64, root int, comm C.MPI_Comm) {
	err := C.MPI_Gather(unsafe.Pointer(&send[0]), C.int(len(send)),
		INT64, unsafe.Pointer(&recv[0]), C.int(len(recv)),
		INT64, C.int(root), comm)
	processError(err)
}

func Scatter_int64(send, recv []int64, root int, comm C.MPI_Comm) {
	err := C.MPI_Scatter(unsafe.Pointer(&send[0]), C.int(len(send)),
		INT64, unsafe.Pointer(&recv[0]), C.int(len(recv)),
		INT64, C.int(root), comm)
	processError(err)
}

func Alltoallv_int64(send []int64, sendCounts, sendDisp []int,
	recv[]int64, recvCounts, recvDisp []int, comm C.MPI_Comm) {

	// Converting between Go and C pointers is way easier if we just do this.
	// It doesn't have any impact on correctness since index [0] isn't used.
	if len(send) == 0 { send = []int64{ 0 } }
	if len(recv) == 0 { recv = []int64{ 0 } }
	
	n := len(sendCounts)
	cSendCounts := make([]C.int, n)
	cSendDisp := make([]C.int, n)
	cRecvCounts := make([]C.int, n)
	cRecvDisp := make([]C.int, n)

	for i := range sendCounts {
		cSendCounts[i], cSendDisp[i] = C.int(sendCounts[i]), C.int(sendDisp[i])
		cRecvCounts[i], cRecvDisp[i] = C.int(recvCounts[i]), C.int(recvDisp[i])
	}

	rank := Comm_rank(COMM_WORLD)
	
	fmt.Println(rank, len(send), len(cSendCounts), len(cSendDisp))
	fmt.Println(rank, len(recv), len(cRecvCounts), len(cRecvDisp))
	err := C.MPI_Alltoallv(unsafe.Pointer(&send[0]), &cSendCounts[0],
		&cSendDisp[0], INT64, unsafe.Pointer(&recv[0]), &cRecvCounts[0],
		&cRecvDisp[0], INT64, comm)
	processError(err)
}

func Alltoallv_int32(send []int32, sendCounts, sendDisp []int,
	recv[]int32, recvCounts, recvDisp []int, comm C.MPI_Comm) {

	// Converting between Go and C pointers is way easier if we just do this.
	// It doesn't have any impact on correctness since index [0] isn't used.
	if len(send) == 0 { send = []int32{ 0 } }
	if len(recv) == 0 { recv = []int32{ 0 } }
	
	n := len(sendCounts)
	cSendCounts := make([]C.int, n)
	cSendDisp := make([]C.int, n)
	cRecvCounts := make([]C.int, n)
	cRecvDisp := make([]C.int, n)

	for i := range sendCounts {
		cSendCounts[i], cSendDisp[i] = C.int(sendCounts[i]), C.int(sendDisp[i])
		cRecvCounts[i], cRecvDisp[i] = C.int(recvCounts[i]), C.int(recvDisp[i])
	}

	rank := Comm_rank(COMM_WORLD)
	
	fmt.Println(rank, len(send), len(cSendCounts), len(cSendDisp))
	fmt.Println(rank, len(recv), len(cRecvCounts), len(cRecvDisp))
	err := C.MPI_Alltoallv(unsafe.Pointer(&send[0]), &cSendCounts[0],
		&cSendDisp[0], INT32, unsafe.Pointer(&recv[0]), &cRecvCounts[0],
		&cRecvDisp[0], INT32, comm)
	processError(err)
}

func Alltoallv_float32(send []float32, sendCounts, sendDisp []int,
	recv[]float32, recvCounts, recvDisp []int, comm C.MPI_Comm) {

	// Converting between Go and C pointers is way easier if we just do this.
	// It doesn't have any impact on correctness since index [0] isn't used.
	if len(send) == 0 { send = []float32{ 0 } }
	if len(recv) == 0 { recv = []float32{ 0 } }
	
	n := len(sendCounts)
	cSendCounts := make([]C.int, n)
	cSendDisp := make([]C.int, n)
	cRecvCounts := make([]C.int, n)
	cRecvDisp := make([]C.int, n)

	for i := range sendCounts {
		cSendCounts[i], cSendDisp[i] = C.int(sendCounts[i]), C.int(sendDisp[i])
		cRecvCounts[i], cRecvDisp[i] = C.int(recvCounts[i]), C.int(recvDisp[i])
	}

	rank := Comm_rank(COMM_WORLD)
	
	fmt.Println(rank, len(send), len(cSendCounts), len(cSendDisp))
	fmt.Println(rank, len(recv), len(cRecvCounts), len(cRecvDisp))
	err := C.MPI_Alltoallv(unsafe.Pointer(&send[0]), &cSendCounts[0],
		&cSendDisp[0], FLOAT32, unsafe.Pointer(&recv[0]), &cRecvCounts[0],
		&cRecvDisp[0], FLOAT32, comm)
	processError(err)
}

func Alltoallv_float64(send []float64, sendCounts, sendDisp []int,
	recv[]float64, recvCounts, recvDisp []int, comm C.MPI_Comm) {

	// Converting between Go and C pointers is way easier if we just do this.
	// It doesn't have any impact on correctness since index [0] isn't used.
	if len(send) == 0 { send = []float64{ 0 } }
	if len(recv) == 0 { recv = []float64{ 0 } }
	
	n := len(sendCounts)
	cSendCounts := make([]C.int, n)
	cSendDisp := make([]C.int, n)
	cRecvCounts := make([]C.int, n)
	cRecvDisp := make([]C.int, n)

	for i := range sendCounts {
		cSendCounts[i], cSendDisp[i] = C.int(sendCounts[i]), C.int(sendDisp[i])
		cRecvCounts[i], cRecvDisp[i] = C.int(recvCounts[i]), C.int(recvDisp[i])
	}

	rank := Comm_rank(COMM_WORLD)
	
	fmt.Println(rank, len(send), len(cSendCounts), len(cSendDisp))
	fmt.Println(rank, len(recv), len(cRecvCounts), len(cRecvDisp))
	err := C.MPI_Alltoallv(unsafe.Pointer(&send[0]), &cSendCounts[0],
		&cSendDisp[0], FLOAT64, unsafe.Pointer(&recv[0]), &cRecvCounts[0],
		&cRecvDisp[0], FLOAT64, comm)
	processError(err)
}

func main() {
	Init()

	//procs := Comm_size(COMM_WORLD)
	rank := Comm_rank(COMM_WORLD)

	var sendCount, sendDisp, recvCount, recvDisp []int
	var send, recv []int64
	switch rank {
	case 0:
		send = []int64{ 110, 111, 120, 130, 131 }
		sendCount = []int{ 2, 1, 2 }
		sendDisp = []int{ 0, 2, 3 }

		recv = []int64{ 0, 0, 0, 0 }
		recvCount = []int{ 2, 0, 2 }
		recvDisp = []int{ 0, 2, 2 }
	case 1:
		send = []int64{ }
		sendCount = []int{ 0, 0, 0 }
		sendDisp = []int{ 0, 0, 0 }

		recv = []int64{ 0 }
		recvCount = []int{ 1, 0, 0 }
		recvDisp = []int{ 0, 1, 1 }
	case 2:
		send = []int64{ 310, 311, 330, 331, 333 }
		sendCount = []int{ 2, 0, 3 }
		sendDisp = []int{ 0, 2, 2 }

		recv = []int64{ 0, 0, 0, 0, 0 }
		recvCount = []int{ 2, 0, 3 }
		recvDisp = []int{ 0, 2, 2 }
	}

	Alltoallv_int64(send, sendCount, sendDisp,
		recv, recvCount, recvDisp, COMM_WORLD)

	fmt.Println(recv)
	
	Finalize()
}
