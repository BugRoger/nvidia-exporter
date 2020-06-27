DATE     = $(shell date +%Y%m%d%H%M)
IMAGE    ?= mrpk1906/nvidia-exporter
VERSION  := $(DATE)
GOOS     ?= $(shell go env | grep GOOS | cut -d'"' -f2)
BINARIES := nvidia-exporter

LDFLAGS := -X github.com/mrpk1906/nvidia-exporter/main.VERSION=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS)"

SRCDIRS  := .
PACKAGES := $(shell find $(SRCDIRS) -type d)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: all clean

all: $(BINARIES:%=bin/$(GOOS)/%)

bin/%: $(GOFILES) Makefile
	GOOS=$(*D) GOARCH=amd64 go build $(GOFLAGS) -v -i -o $(@D)/$(@F) .

build: 
	docker build $(BUILD_ARGS) -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

push: build
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

clean:
	rm -rf bin/*

