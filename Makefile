.PHONY: pre-pr
pre-pr: fmt-check vet test

.PHONY: fmt
fmt:
	@gofmt -w .

.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

.PHONY: vet
vet:
	@go vet ./...

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
