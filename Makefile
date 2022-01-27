# BUILD_VERSION is the latest tag.
BUILD_VERSION := $(shell git describe --tags --always)

#
# Go sources management.
#

GO := $(shell which go)

.PHONY: build
build:
	$(GO) build -ldflags "-X github.com/optable/match-cli/pkg/cli.version=${BUILD_VERSION}" -o bin/match-cli cmd/cli/main.go

.PHONY: release
release: darwin linux windows

.PHONY: darwin
darwin: darwin-amd64 darwin-arm64
	$(GO) install github.com/randall77/makefat@latest
	makefat release/match-cli-darwin release/match-cli-darwin-amd64 release/match-cli-darwin-arm64
	rm release/match-cli-darwin-amd64 release/match-cli-darwin-arm64

.PHONY: darwin-amd64
darwin-amd64:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 make build ;\
	mkdir -p release && cp bin/match-cli release/match-cli-darwin-amd64

.PHONY: darwin-arm64
darwin-arm64:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 make build ;\
	mkdir -p release && cp bin/match-cli release/match-cli-darwin-arm64

.PHONY: linux
linux:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build ;\
	mkdir -p release && cp bin/match-cli release/match-cli-linux-amd64

.PHONY: windows
windows:
	make clean-bin ;\
 	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 make build ;\
	mkdir -p release && cp bin/match-cli release/match-cli-windows-amd64.exe

.PHONY: clean
clean: clean-bin clean-release

.PHONY: clean-bin
clean-bin:
	rm -f ./bin/match-cli

.PHONY: clean-release
clean-release:
	rm -f release/*
