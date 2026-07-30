.PHONY: all build test clean run

all: build

build:
	go build -o sd .

test:
	go test ./...

clean:
	go clean
	rm -f sd

run: build
	cd example && ../sd
