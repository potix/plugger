CDIR := $(shell pwd)

all:
	glide update
	GOPATH=${GOPATH}:$(CDIR) go build -buildmode=c-shared -o sample2.so sample2.go
install:
	glide install
