export GOBIN := ${PWD}/bin
export PATH  := ${GOBIN}:${PATH}

BINARY_NAME = 'goscout'
BINARY_PATH = ${GOBIN}/${BINARY_NAME}

# all src packages without generated code
PACKAGES = $(shell go list ./...)

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
	@echo "[cleaning]"
	@go clean
	@if [ -f ${BINARY_PATH} ] ; then rm -v ${BINARY_PATH} ; fi
	@rm -rfv ${GOBIN}

download-deps:
	@echo "[download dependencies]"
	@go mod download
	@go mod download -modfile=tools/go.mod

tidy:
	@echo "[tidying]"
	@go mod tidy

generate:
	@echo "[generate]"
	@go install -modfile=tools/go.mod github.com/golang/mock/mockgen
	@find . -not -path '*/\.*' -name \*_mock.go -delete
	@go generate ${PACKAGES}

lint:
	@echo "[lint]"
	@go install -modfile=tools/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint
	@${GOBIN}/golangci-lint run

test:
	@echo "[test]"
	@go test -race -v -count=1 ./...

build:
	@echo "[build]"
	@go build -a -o ${BINARY_PATH}
	@echo "Compiled successfully!"
	@echo "Output directory: ${GOBIN}"
