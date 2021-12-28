WORKDIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

APP := noso-go
BINARIES := $(APP)-linux-amd64 $(APP)-linux-386 $(APP)-darwin-amd64 $(APP)-windows-386 $(APP)-windows-amd64 $(APP)-linux-arm $(APP)-linux-arm64

REVISION := $(shell git rev-parse --short=8 HEAD)
TAG := $(shell git describe --tags --exact-match $(REVISION) 2>/dev/null)

GO111MODULE ?= auto
GOFLAGS ?= -mod=vendor

ifneq ($(TAG),)
VER := $(TAG)
LDVER := -X 'github.com/Noso-Project/noso-go/internal/miner.Version=$(VER)'
else
LDVER :=
endif

LDREV := -X 'github.com/Noso-Project/noso-go/internal/miner.Commit=$(REVISION)'

LDFLAGS := -ldflags "-s -w $(LDVER) $(LDREV)"

.PHONY: patch
patch: patch-$(APP)-windows-386 patch-$(APP)-windows-amd64

.PHONY: patch-%
patch-%:
	@if test $(findstring windows,$(subst -, ,$*)); then \
		echo "Patching icon into Windows binary: $*"; \
		go-winres patch bin/$*; \
	else \
		echo "Wont patch icon into non-Windows binary: $*"; \
	fi;

.PHONY: all
all: $(BINARIES)

.PHONY: $(APP)-%
$(APP)-%: OS=$(word 3,$(subst -, ,$@))
$(APP)-%: ARCH=$(word 4,$(subst -, ,$@))
$(APP)-%:
	$(info # #########################################################)
	$(info #)
	$(info # Building $(APP) binary for OS $(OS) and Arch $(ARCH))
	$(info #)
	$(info # #########################################################)
	GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/$@ $(LDFLAGS) main.go

.PHONY: packages
packages: package-linux-386 package-linux-amd64 package-linux-arm package-linux-arm64 package-darwin-amd64 package-windows-386 package-windows-amd64

.PHONY: package-%
package-%: OS=$(word 2,$(subst -, ,$@))
package-%: ARCH=$(word 3,$(subst -, ,$@))
package-%:
	$(info # #########################################################)
	$(info #)
	$(info # Packaging $(APP) binary for OS $(OS) and Arch $(ARCH))
	$(info #)
	$(info # #########################################################)
	@cp README.md bin/
	@case $(OS) in \
		linux) \
			cp bin/$(APP)-$(OS)-$(ARCH) bin/$(APP);\
			cp examples/noso-go.sh bin/noso-go.sh;\
			chmod +x bin/$(APP);\
			chmod +x examples/noso-go.sh;\
			(cd bin && tar -zcvf ../packages/$(APP)-$(TAG)-$(OS)-$(ARCH).tgz $(APP) README.md noso-go.sh); \
			(cd bin && zip ../packages/$(APP)-$(TAG)-$(OS)-$(ARCH).zip $(APP) README.md noso-go.sh); \
			;; \
		darwin) \
			cp bin/$(APP)-$(OS)-$(ARCH) bin/$(APP);\
			cp examples/noso-go.sh bin/noso-go.sh;\
			chmod +x bin/$(APP);\
			chmod +x examples/noso-go.sh;\
			(cd bin && zip ../packages/$(APP)-$(TAG)-$(OS).zip $(APP) README.md noso-go.sh); \
			;; \
		windows) \
			cp bin/$(APP)-$(OS)-$(ARCH) bin/$(APP).exe;\
			cp examples/$(APP).bat bin/$(APP).bat;\
			cp examples/$(APP)-menu.bat bin/$(APP)-menu.bat;\
			chmod +x bin/$(APP).exe;\
			if [ "$(ARCH)" = "386" ]; then \
				(cd bin && zip ../packages/$(APP)-$(TAG)-win32.zip $(APP).exe $(APP).bat ${APP}-menu.bat README.md); \
			else \
				(cd bin && zip ../packages/$(APP)-$(TAG)-win64.zip $(APP).exe $(APP).bat ${APP}-menu.bat README.md); \
			fi \
			;; \
	esac

ifeq (, $(shell which richgo))
gotest := go test
else
gotest := richgo test
endif

.PHONY: unit-test
unit-tests:
	$(gotest) -v -race -cover -timeout 10s ./...

.PHONY: benchmark-%
benchmark-%:
	#  -run=XXX excludes unit tests
	#  -bench "(?i)$*$$" is a regex match
	#         (?i) makes it case insensitive
	#         $* matches the % in benchmark-%
	#         $$ is make's version of an escaped $
	$(gotest) -run=XXX -bench "(?i)$*$$" -benchtime 5s -v -race  ./... -cpu=1,2

.PHONY: clean
clean:
	rm -f bin/*
	rm -f packages/*
