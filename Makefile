GOPATH := $(PWD)
GOBIN  := $(GOPATH)/bin

dep:
	go get

build:
	go build

clean:
	go clean
	rm -f $(GOBIN)/*
