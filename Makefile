Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/debug/universal-exporter cmd/*.go

.PHONY: release
release:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -o build/release/universal-exporter-linux cmd/*.go
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LDFLAGS) -o build/release/universal-exporter-darwin cmd/*.go
