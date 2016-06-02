CDIR := $(shell pwd)

all:
	glide update
	GOPATH=${GOPATH}:$(CDIR) go build -o core
install:
	glide install
