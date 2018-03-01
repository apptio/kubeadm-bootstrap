# VERSION is the last git tag
# e.g. 0.1
VERSION = $(shell git describe --abbrev=0)

# GOVERSION is the current go version
# e.g. go1.10
GOVERSION = $(shell go version | awk '{print $$3;}')

# GORELEASER is the path to the goreleaser binary.
# e.g. /usr/local/bin/goreleaser
GORELEASER = $(shell which goreleaser)

.PHONY: clean build

build: clean
	@[ -x "$(GORELEASER)" ] || ( echo "goreleaser not installed"; exit 1)
	@GOVERSION=$(GOVERSION) goreleaser --rm-dist

clean:
	@rm -rf dist pkg kubeadm-bootstrap
