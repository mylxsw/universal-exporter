Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/universal-exporter cmd/*.go

.PHONY: dist
dist:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/universal-exporter-linux cmd/*.go
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/universal-exporter-darwin cmd/*.go
