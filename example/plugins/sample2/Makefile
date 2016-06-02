CDIR := $(shell pwd)

all:
	glide update
	go build -buildmode=c-shared -o sample2.so sample2.go
install:
	glide install
clean:
	rm -rf sample2.so glide.lock vendor
