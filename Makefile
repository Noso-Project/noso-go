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
all: $(APP)-linux-x86_64 $(APP)-linux-i386 $(APP)-darwin $(APP)-win32.exe $(APP)-win64.exe $(APP)-arm $(APP)-arm64

$(APP)-linux-x86_64:
	GOOS=linux GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-linux-i386:
	GOOS=linux GOARCH=386 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-win64.exe:
	GOOS=windows GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-win32.exe:
	GOOS=windows GOARCH=386 go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-arm:
	GOOS=linux GOARCH=arm go build -o $@ $(LDFLAGS) cmd/miner/main.go

$(APP)-arm64:
	GOOS=linux GOARCH=arm64 go build -o $@ $(LDFLAGS) cmd/miner/main.go

.PHONY: package-%
package-%: OS=$(word 2,$(subst -, ,$@))
package-%: ARCH=$(word 3,$(subst -, ,$@))
package-%:
	mkdir -p packages;\
	case $(OS) in \
		linux) \
			cp bin/$(APP)-$(OS)-$(ARCH) bin/$(APP);\
			chmod +x bin/$(APP);\
			(cd bin && tar -zcvf ../$(APP)-$(TAG)-$(OS)-$(ARCH).tgz $(APP)); \
			;; \
		darwin) \
			cp bin/$(APP)-$(OS) bin/$(APP);\
			chmod +x bin/$(APP);\
			(cd bin && zip ../$(APP)-$(TAG)-$(OS).zip $(APP)); \
			;; \
		windows) \
			cp bin/$(APP)-$(OS) bin/$(APP);\
			chmod +x bin/$(APP);\
			(cd bin && zip ../$(APP)-$(TAG)-$(OS).zip $(APP)); \
			;; \
	esac

.PHONY: clean
clean:
	rm -f $(APP)-*.exe
	rm -f $(APP)-linux*
	rm -f $(APP)-darwin*
	rm -f $(APP)-arm*
	rm -f packages/*.tgz
	rm -f packages/*.zip
