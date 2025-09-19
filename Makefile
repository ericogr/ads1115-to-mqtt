.PHONY: all build build-dietpi build-linux build-windows clean

BINARY=ads1115-to-mqtt

all: build

RUN_ARGS ?= -outputs console -sensor-type simulation

run:
	go run ./ $(RUN_ARGS)


build:
	mkdir -p bin
	go build -o bin/$(BINARY) -v ./

build-dietpi:
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
