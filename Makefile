BINARY_NAME = goscout

BASEPATH = $(shell pwd)

export GOBIN := $(BASEPATH)/bin

PATH := $(GOBIN):$(PATH)

# tools
LINTER = $(GOBIN)/golangci-lint

# binary path
BINARY_PATH = $(GOBIN)/$(BINARY_NAME)

# all src packages without generated code
PKGS = $(shell go list ./...)

help:
	@echo 'Usage: make <TARGETS> ... <OPTIONS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help'
	@echo '    clean              Remove binaries'
	@echo '    download-deps      Download and install dependencies'
	@echo '    tidy               Perform go tidy steps'
	@echo '    generate           Perform go generate'
	@echo '    lint               Run all linters'
	@echo '    test               Run unit tests'
	@echo '    build              Compile packages and dependencies'
	@echo ''

clean:
	@go clean
	@if [ -f $(BINARY_PATH) ] ; then rm $(BINARY_PATH) ; fi
	@rm -rf $(GOBIN)

download-deps:
	@go mod download
	@go mod download -modfile=tools/go.mod

tidy:
	@go mod tidy

generate:
	@go install -modfile=tools/go.mod github.com/golang/mock/mockgen
	@find . -not -path '*/\.*' -name \*_mock.go -delete
	@go generate $(PKGS)

lint:
	@go install -modfile=tools/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint
	@$(LINTER) run

test:
	@go test -race -v -count=1 ./...

build:
	@CGO_ENABLED=0 go build -a -installsuffix cgo -o $(BINARY_PATH)
