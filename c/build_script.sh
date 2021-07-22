go build -buildmode=c-archive go_wrapper.go && gcc -Wall -Wextra -pthread read_guppy.c go_wrapper.a -o main && exit 0
exit 1
