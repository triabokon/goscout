BINARY_NAME = goscout

BASEPATH = $(shell pwd)

export GOBIN := $(BASEPATH)/bin

PATH := $(GOBIN):$(PATH)

# Basic go commands
GOBUILD    = go build
GOINSTALL  = go install
GOCLEAN    = go clean
GOTEST     = go test
GODOWNLOAD = go mod download
GOTIDY     = go mod tidy

# tools
LINTER = $(GOBIN)/golangci-lint

# binary path
BINARY_PATH = $(GOBIN)/$(BINARY_NAME)

help:
	@echo 'Usage: make <TARGETS> ... <OPTIONS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help'
	@echo '    clean              Remove binaries'
	@echo '    download-deps      Download and install build time dependencies'
	@echo '    tidy               Perform go tidy steps'
	@echo '    lint               Run all linters including vet and gosec and others'
	@echo '    test               Run unit tests'
	@echo '    build              Compile packages and dependencies'
	@echo ''

clean:
	@$(GOCLEAN)
	@if [ -f $(BINARY_PATH) ] ; then rm $(BINARY_PATH) ; fi
	@rm -rf $(GOBIN)

download-deps:
	@$(GODOWNLOAD)
	@$(GODOWNLOAD) -modfile=tools/go.mod

tidy:
	@$(GOTIDY)

lint:
	@$(GOINSTALL) -modfile=tools/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint
	@$(LINTER) run

test:
	@$(GOTEST) -race -v -count=1 ./...

build:
	@CGO_ENABLED=0 $(GOBUILD) -ldflags $(BUILD_LDFLAGS) -a -installsuffix cgo -o $(BINARY_PATH)