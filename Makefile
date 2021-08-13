# BUILD_VERSION is the version of the build.
BUILD_VERSION="${CIRCLE_TAG}"
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
	$(GO) build -ldflags "-X github.com/optable/match-cli/pkg/cli.version=${BUILD_VERSION}" -o $@ $<

.PHONY: build
build: $(CLI_BIN)

.PHONY: release
release: darwin linux windows

.PHONY: darwin
darwin:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 make bin/match-cli ;\
	mkdir -p release && cp bin/match-cli release/match-cli-darwin-amd64

.PHONY: linux
linux:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make bin/match-cli ;\
	mkdir -p release && cp bin/match-cli release/match-cli-linux-amd64

.PHONY: windows
windows:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 make bin/match-cli ;\
	mkdir -p release && cp bin/match-cli release/match-cli-windows-amd64.exe

.PHONY: clean
clean: clean-bin clean-release
	
.PHONY: clean-bin
clean-bin: 
	rm -f $(CLI_BIN)
	
.PHONY: clean-release
clean-release:
	rm -f release/*
