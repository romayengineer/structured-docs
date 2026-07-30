.PHONY: all build test clean run cover cover-html

all: build

build:
	go build -o sd .

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

run: build
	cd example && ../sd
