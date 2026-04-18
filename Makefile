.PHONY: pre-pr
pre-pr: test

GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

bin/litdoc: $(GO_FILES)
	@go build -o bin/litdoc .

.PHONY: build
build: bin/litdoc

.PHONY: test
test: build
	@go test ./... --count=1

.PHONY: clean
clean:
	@rm -rf bin/