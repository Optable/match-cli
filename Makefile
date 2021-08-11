# BUILD_VERSION is the version of the build.
#BUILD_VERSION := $(shell git describe)
BUILD_VERSION = 1.0.0
# BUILD_COMMIT is the commit from which the binary was build.
BUILD_COMMIT := $(shell git rev-parse HEAD)
# BUILD_DATE is the date at which the binary was build.
BUILD_DATE := $(shell date -u "+%Y-%m-%dT%H:%M:%S+00:00")

#
# Go sources management.
#

GO := $(shell which go)

CLI_CMD = bin/match-cli
CLI_BIN = $(subst cmd,bin,$(CLI_CMD))
CLI_FILES := $(SRC_FILES) $(COMMON_SRC_FILES) $(CLIENT_SRC_FILES)

bin/match-cli: cmd/cli/main.go $(CLI_FILES)
	$(GO) build -o $@ $<

.PHONY: build
build: $(CLI_BIN)

.PHONY: release
release: darwin linux windows

.PHONY: darwin
darwin:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 make bin/match-cli ;\
	mkdir -p release && cp bin/match-cli release/match-cli-darwin-amd64-$(BUILD_VERSION)

.PHONY: linux
linux:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make bin/match-cli ;\
	mkdir -p release && cp bin/match-cli release/match-cli-linux-amd64-$(BUILD_VERSION)

.PHONY: windows
windows:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 make bin/match-cli ;\
	mkdir -p release && cp bin/match-cli release/match-cli-windows-amd64-$(BUILD_VERSION).exe

.PHONY: clean
clean: clean-bin clean-release
	
.PHONY: clean-bin
clean-bin: 
	rm -f $(CLI_BIN)
	
.PHONY: clean-release
clean-release:
	rm -f release/*
