BINARY := CFDNSU
CONFIG_FILE := cfdnsu.conf

GOPATH := $(PWD)
GOBIN := $(GOPATH)/bin
SYSCONFDIR := $(PREFIX)/etc
CONFIG_PATH := $(SYSCONFDIR)/$(CONFIG_FILE)
SBINDIR := $(PREFIX)/usr/sbin
LIBDIR := $(PREFIX)/usr/lib
PLUGINS := $(wildcard plugin/*.go)

all : build build-plugins

build-plugins : $(PLUGINS)
	$(foreach plugin,$^, go build -buildmode=plugin -ldflags "-s" -o $(GOBIN)/$(notdir $(patsubst %.go,%.so,$(plugin))) $(plugin);)

build : main.go
	go build -ldflags "-s -X main.CONFIGURATION_PATH=${CONFIG_PATH}" -o $(GOBIN)/$(BINARY)

dep:
	go get -d

.PHONY: all build build-plugins dep
#install:
#	mkdir -p $(SYSCONFDIR) $(SBINDIR)
#	cp $(CONFIG_FILE).template $(CONFIG_PATH)
#	cp $(GOBIN)/$(BINARY) $(SBINDIR)/$(BINARY)

#clean:
#	go clean
#	rm -rf $(GOBIN)/*
