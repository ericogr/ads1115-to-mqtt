
.PHONY: all build build-dietpi build-linux build-windows clean

BINARY=ads1115-to-mqtt

# Docker image settings (override as needed)
DOCKER_IMAGE ?= ericogr/ads1115-to-mqtt
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
VERSION ?= latest
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo '')
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

all: build

RUN_ARGS ?= ""

run:
	go run ./ $(RUN_ARGS)


build:
	mkdir -p bin
	go build -o bin/$(BINARY) -v ./

build-linux-arm64:
	mkdir -p bin
	GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY)-linux-arm64 -v ./

build-linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY)-linux-amd64 -v ./

build-windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY)-windows-amd64.exe -v ./

clean:
	rm -rf bin/*

.PHONY: test
test:
	GOCACHE=/tmp/gocache go test ./...

# Docker related targets
.PHONY: docker-build docker-buildx
docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg BUILD_DATE=$(BUILD_DATE) -t $(DOCKER_IMAGE):$(VERSION) .

docker-buildx:
	@echo "Building multi-platform image for $(DOCKER_PLATFORMS)"
	docker buildx build --platform $(DOCKER_PLATFORMS) --push \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest \
		.
