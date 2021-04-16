WORKDIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

APP := go-miner

REVISION := $(shell git rev-parse --short=8 HEAD)
TAG := $(shell git describe --tags --exact-match $(REVISION) 2>/dev/null)

ifneq ($(TAG),)
VER := $(TAG)
else
VER := commit-$(REVISION)
endif

LDFLAGS := -ldflags "-X 'github.com/leviable/noso-go/internal/miner.Version=$(VER)'"

.PHONY: all
all: $(APP)-linux $(APP)-macos $(APP).exe $(APP)-arm $(APP)-arm64

$(APP)-linux:
	GOOS=linux GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-macos:
	GOOS=darwin GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP).exe:
	GOOS=windows GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-arm:
	GOOS=linux GOARCH=arm go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-arm64:
	GOOS=linux GOARCH=arm64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-$(TAG).zip:
	(cd binaries && zip ../$@ ./*)

.PHONY: zip
zip: $(APP)-$(TAG).zip

.PHONY: clean
clean:
	rm -f $(APP).exe
	rm -f $(APP)-linux
	rm -f $(APP)-macos
	rm -f $(APP)-arm
	rm -f $(APP)-arm64
