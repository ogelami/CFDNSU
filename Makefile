GOPATH := $(PWD)
GOBIN  := $(GOPATH)/bin
BINARY := CFDNSU
CONFIG_PATH := /etc/nowhere/fun.conf

dep:
	go get

build:
	go build -ldflags "-X main.CONFIGURATION_PATH=${CONFIG_PATH}" -o $(GOBIN)/$(BINARY)

clean:
	go clean
	rm -f $(GOBIN)/*
