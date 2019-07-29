SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
GIT_COMMIT := $(shell git rev-parse HEAD)
GO_PKGS := $(shell go list ./...)
GO_LDFLAGS := "-X github.com/romantomjak/b2/version.GitCommit=$(GIT_COMMIT)"

.PHONY: build
build:
	go build -ldflags $(GO_LDFLAGS) -o b2

.PHONY: test
test:
	go test -cover $(GO_PKGS)

.PHONY: clean
clean:
	@rm -f "$(PROJECT_ROOT)/b2"
