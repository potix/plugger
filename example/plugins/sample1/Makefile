CDIR := $(shell pwd)

all:
	glide update
	go build -buildmode=c-shared -o sample1.so sample1.go
install:
	glide install
clean:
	rm -rf sample1.so glide.lock vendor
