.PHONY: build test clean

build:
	go build -o bin/litdoc .

test: build
	go test ./... --count=1

clean:
	rm -rf bin/