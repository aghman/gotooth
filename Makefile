.PHONY: struct build build_rp build_mac build_linux clean

struct: 
	- @mkdir build

build: clean struct build_rp build_mac build_linux

build_rp:
	GOOS=linux GOARCH=arm GOARM=5 go build -o ${PWD}/build/gotooth.rp

build_mac:
	GOOS=darwin go build -o ${PWD}/build/gotooth.darwin

build_linux:
	GOOS=linux go build -o ${PWD}/build/gotooth.linux

clean:
	-rm -rf build