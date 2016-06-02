CDIR := $(shell pwd)

all:
	glide update
	go build -o core
install:
	glide install
clean:
	rm -rf core glide.lock vendor
