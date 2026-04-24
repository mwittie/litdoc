.PHONY: pre-pr
pre-pr: test

.PHONY: mock
mock:
	@mockery

.PHONY: mock-clean
mock-clean:
	@find . \( -name '*_mock_test.go' -o -name '*_mock.go' \) -not -path './vendor/*' -delete

GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')
GOCACHE ?= /tmp/litdoc-go-build

bin/litdoc: $(GO_FILES)
	@GOCACHE=$(GOCACHE) go build -o bin/litdoc .

.PHONY: build
build: bin/litdoc

.PHONY: test
test: build
	@GOCACHE=$(GOCACHE) go test ./... --count=1

.PHONY: clean
clean:
	@rm -rf bin/
