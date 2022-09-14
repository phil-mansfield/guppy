# Install the Go library used as the base of the C library
go install github.com/phil-mansfield/guppy/go &&
	# Compile the Go wrapper wound the Go library into an archive file
	go build -buildmode=c-archive guppy_wrapper.go &&
	# Rename the archive file to the format the C compiler expects
	mv guppy_wrapper.a libguppy_wrapper.a &&
	# Compile the C wrapper around the Go code into an object file
	gcc -c -L. -std=c99 -Wall -Wextra -O2 read_guppy.c -pthread -lguppy_wrapper &&
	# Convert the object file into an archive file
	ar -r libread_guppy.a read_guppy.o &&
	# Compile the test file into an executable binary
	gcc -L. -std=c99 read_guppy_test.c -lread_guppy -lguppy_wrapper -lpthread -o read_guppy_test &&
	# Exit normally
	exit 0
# Uh oh, there's a problem
exit 1
