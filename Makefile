BINARY_NAME=ht
VERSION?=0.1.0

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

LD_FLAGS=-ldflags "-X github.com/hackerspace/hacktrackmmu-cli/internal/version.Version=$(VERSION)"

.PHONY: all build clean test fmt tidy install run

all: build

build:
	mkdir -p bin
	$(GOBUILD) $(LD_FLAGS) -o bin/$(BINARY_NAME) main.go

run:
	$(GOCMD) run $(LD_FLAGS) main.go $(ARGS)

clean:
	$(GOCLEAN)
	rm -rf bin

test:
	$(GOTEST) -v ./...

fmt:
	$(GOFMT) -s -w .

tidy:
	$(GOMOD) tidy

install:
	$(GOCMD) install $(LD_FLAGS)
