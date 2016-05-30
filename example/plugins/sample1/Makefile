CDIR := $(shell pwd)

all:
	glide update
	ln -sf vendor src
	GOPATH=${GOPATH}:$(CDIR) go build -buildmode=c-shared -o sample1.so sample1.go
install:
	glide install
