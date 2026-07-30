.PHONY: all build test clean run cover cover-html install

all: build

build:
	go build -o sd ./cmd/sd

test:
	go test ./...

clean:
	go clean
	rm -f sd

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out -o coverage.html

install: build
	mkdir -p ~/.local/bin
	cp sd ~/.local/bin/sd

run: build
	cd example && ../sd
