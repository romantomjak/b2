SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
VERSION := 0.4.0
GIT_COMMIT := $(shell git rev-parse --short HEAD)

GO_PKGS := $(shell go list ./...)
GO_LDFLAGS := "-X github.com/romantomjak/b2/version.Version=$(VERSION) -X github.com/romantomjak/b2/version.GitCommit=$(GIT_COMMIT)"

PLATFORMS := darwin linux windows
os = $(word 1, $@)

.PHONY: build
build:
	@mkdir -p $(PROJECT_ROOT)/bin
	@go build -ldflags $(GO_LDFLAGS) -o $(PROJECT_ROOT)/bin/b2

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	@mkdir -p $(PROJECT_ROOT)/dist
	@GOOS=$(os) GOARCH=amd64 go build -o $(PROJECT_ROOT)/dist/$(os)/b2 github.com/romantomjak/b2
	@zip -q -X -j $(PROJECT_ROOT)/dist/b2_$(VERSION)_$(os)_amd64.zip $(PROJECT_ROOT)/dist/$(os)/b2
	@rm -rf $(PROJECT_ROOT)/dist/$(os)

.PHONY: release
release: windows linux darwin

.PHONY: test
test:
	@go test -cover $(GO_PKGS)

.PHONY: clean
clean:
	@rm -rf "$(PROJECT_ROOT)/bin/"
	@rm -rf "$(PROJECT_ROOT)/dist/"
