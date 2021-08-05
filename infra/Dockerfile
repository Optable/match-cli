FROM golang:1.16-alpine3.13 AS build

RUN apk update \
  && apk add --no-cache bash git make protobuf-dev protoc curl

WORKDIR /build
RUN mkdir publish

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0 && \
    git clone --depth 1 https://github.com/Optable/match.git

COPY ./go.mod ./go.sum /build/

COPY ./pkg /build/pkg/
COPY ./internal /build/internal/
COPY ./api /build/api
COPY ./cmd /build/cmd

COPY Makefile /build/Makefile

COPY ./infra/run.sh .
RUN chmod +x ./run.sh
RUN ./run.sh --build publish

FROM google/cloud-sdk:alpine AS publish

COPY --from=build /build/publish /publish
WORKDIR /publish
ENV PATH="/publish:$PATH"
RUN curl -o semver "https://raw.githubusercontent.com/fsaintjacques/semver-tool/3.0.0/src/semver" && chmod +x semver
COPY ./infra/run.sh .
RUN chmod +x ./run.sh
