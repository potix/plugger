CDIR := $(shell pwd)

all:
	glide update
	ln -sf vendor src
	GOPATH=${GOPATH}:$(CDIR) go build -o core
install:
	glide install
