BINARY := CFDNSU
CONFIG_FILE := cfdnsu.conf

GOPATH := $(PWD)
GOBIN  := $(GOPATH)/bin
SYSCONFDIR := $(PREFIX)/etc
CONFIG_PATH := $(SYSCONFDIR)/$(CONFIG_FILE)
SBINDIR := $(PREFIX)/usr/sbin

all:
	echo "${PREFIX} ${SYSCONFDIR} ${sbindir} ${SBINDIR}"
dep:
	go get -d

build:
	go build -ldflags "-X main.CONFIGURATION_PATH=${CONFIG_PATH}" -o $(GOBIN)/$(BINARY)
install:
	mkdir -p $(SYSCONFDIR) $(SBINDIR)
	cp $(CONFIG_FILE).template $(CONFIG_PATH)
	cp $(GOBIN)/$(BINARY) $(SBINDIR)/$(BINARY)

clean:
	go clean
	rm -f $(GOBIN)/*
