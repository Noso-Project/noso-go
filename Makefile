WORKDIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

APP := noso-go

REVISION := $(shell git rev-parse --short=8 HEAD)
TAG := $(shell git describe --tags --exact-match $(REVISION) 2>/dev/null)

ifneq ($(TAG),)
VER := $(TAG)
else
VER := commit-$(REVISION)
endif

LDFLAGS := -ldflags "-s -w -X 'github.com/leviable/noso-go/internal/miner.Version=$(VER)'"

.PHONY: all
all: $(APP)-linux-amd64 $(APP)-linux-386 $(APP)-darwin-amd64 $(APP)-windows-386 $(APP)-windows-amd64 $(APP)-linux-arm $(APP)-linux-arm64

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
			chmod +x bin/$(APP).exe;\
			if [ "$(ARCH)" = "386" ]; then \
				(cd bin && zip ../packages/$(APP)-$(TAG)-win32.zip $(APP).exe $(APP).bat README.md); \
			else \
				(cd bin && zip ../packages/$(APP)-$(TAG)-win64.zip $(APP).exe $(APP).bat README.md); \
			fi \
			;; \
	esac

.PHONY: clean
clean:
	rm -f bin/*
	rm -f packages/*
