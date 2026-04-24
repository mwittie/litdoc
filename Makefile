.PHONY: pre-pr
pre-pr: fmt-check vet test

.PHONY: vet
vet:
	@go vet ./...

.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

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