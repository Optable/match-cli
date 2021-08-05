# BUILD_VERSION is the version of the build.
#BUILD_VERSION := $(shell git describe)
# hard code it for now since git describe does not return
BUILD_VERSION = 1.0.0
# BUILD_COMMIT is the commit from which the binary was build.
BUILD_COMMIT := $(shell git rev-parse HEAD)
# BUILD_DATE is the date at which the binary was build.
BUILD_DATE := $(shell date -u "+%Y-%m-%dT%H:%M:%S+00:00")

# DOCKER_TAG controls the tag that is applied to the terminal node (controlled
# with DOCKER_TARGET) in each docker build. For example, one could build an
# image of optable-edge tagged to v1.1.1 with the following call.
#
# ```
# $ make DOCKER_TAG=v1.1.1 docker-build-edge
# $ docker image ls optable-edge:v*
# REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
# optable-edge        v1.1.1              0db544aa55df        2 minutes ago       18.4MB
# ```
DOCKER_TAG ?= $(BUILD_COMMIT)

#
# Go sources management.
#

GO := $(shell which go)

GO_MATCH_CLI_CMD = bin/match-cli
GO_MATCH_CLI_BIN = $(subst cmd,bin,$(GO_MATCH_CLI_CMD))

GO_ADMIN_SRC_FILES := $(shell find internal -type f -name '*.go')
GO_COMMON_SRC_FILES := $(shell find pkg -type f -name '*.go')
GO_MATCH_CLI_SRC_FILES := $(shell find cmd/cli -type f -name '*.go')
GO_MATCH_CLI_FILES := $(GO_MATCH_CLI_SRC_FILES) $(GO_COMMON_SRC_FILES) $(GO_ADMIN_CLIENT_FILES)


.PHONY: publish
publish:
	docker build . --target publish -t match-cli-publish:$(DOCKER_TAG) \
    --file $(shell realpath infra/Dockerfile)
	docker run \
		--volume $(HOME)/.config/gcloud:/root/.config/gcloud \
		match-cli-publish:$(DOCKER_TAG) \
		run.sh --publish gs://optable-cli $(BUILD_VERSION)

bin/match-cli: cmd/cli/main.go $(GO_MATCH_CLI_FILES)
	$(GO) build -o $@ $<

.PHONY: build
build: $(GO_MATCH_CLI_BIN)

.PHONY: clean
clean:
	rm -f $(GO_MATCH_CLI_BIN)

