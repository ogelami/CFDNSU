BINARY := CFDNSU
CONFIG_FILE := cfdnsu.conf

GOPATH := $(PWD)
GOBIN := $(GOPATH)/bin
SYSCONFDIR := $(PREFIX)/etc
CONFIG_PATH := $(SYSCONFDIR)/$(CONFIG_FILE)
SBINDIR := $(PREFIX)/usr/sbin
LIBDIR := $(PREFIX)/usr/lib
PLUGIN_PATH := $(GOPATH)/plugin
PLUGINS := $(wildcard $(PLUGIN_PATH)/*.go)

all: dep build build-plugins
	
dep:
	go get -d

build-plugins: $(PLUGINS)
	go build -buildmode=plugin -o $(GOBIN)/$(patsubst %.go,%.so,$(^F)) $(PLUGIN_PATH)/$(^F)

build:
	go build -ldflags "-X main.CONFIGURATION_PATH=${CONFIG_PATH}" -o $(GOBIN)/$(BINARY)

install:
	mkdir -p $(SYSCONFDIR) $(SBINDIR)
	cp $(CONFIG_FILE).template $(CONFIG_PATH)
	cp $(GOBIN)/$(BINARY) $(SBINDIR)/$(BINARY)

clean:
	go clean
	rm -rf $(GOBIN)/*
