BINARY = handler
CMD = ./cmd/handler
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -s -w -X main.version=$(VERSION)
PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: build test vet lint clean build-all

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(CMD)

test:
	go test -race ./...

vet:
	go vet ./...

lint: vet
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -r bin/ 2>/dev/null || true

build-all: clean
	@$(foreach platform,$(PLATFORMS),\
		$(eval OS=$(word 1,$(subst /, ,$(platform))))\
		$(eval ARCH=$(word 2,$(subst /, ,$(platform))))\
		echo "Building $(OS)/$(ARCH)..." && \
		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
			go build -ldflags "$(LDFLAGS)" \
			-o bin/$(BINARY)-$(OS)-$(ARCH) $(CMD) && \
	) true
