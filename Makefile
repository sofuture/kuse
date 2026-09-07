VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: default
default: build

.PHONY: build
build: build/kuse

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: check
check: vet test

.PHONY: clean
clean:
	rm -rf build dist

build/kuse: $(shell find cmd pkg -type f -name '*.go')
	mkdir -p build
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o build/kuse ./cmd/kuse
