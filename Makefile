.PHONY: test build

VERSION = $(shell cat version.txt)
CURRENT_DIR = $(shell pwd)

build-repo-server:
	go build -o build/repo.server -ldflags "-s -w" cmd/repo.server/main.go

build: build-repo-server
	echo "build done."

test:
	go test ./... && echo ":)" || echo ":-/"
