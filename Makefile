BINARY     := netwatch
MODULE     := github.com/Blizoman/netwatch
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE       := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
	-X '$(MODULE)/internal/version.Version=$(VERSION)' \
	-X '$(MODULE)/internal/version.Commit=$(COMMIT)' \
	-X '$(MODULE)/internal/version.Date=$(DATE)'

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/netwatch

.PHONY: install
install:
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/netwatch

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: cover
cover:
	go test -race -count=1 -coverprofile=coverage.txt ./...
	go tool cover -func=coverage.txt

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	gofmt -l -w .

.PHONY: docker
docker:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t netwatch:$(VERSION) .

.PHONY: clean
clean:
	rm -rf bin dist coverage.txt coverage.html
