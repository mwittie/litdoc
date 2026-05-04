GO ?= go

.PHONY: pre-pr
pre-pr: clean mock fmt-check vet test

.PHONY: fmt
fmt:
	@gofmt -w $(GO_FILES)

.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l $(GO_FILES)); \
	if [ -n "$$unformatted" ]; then \
		echo "unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

.PHONY: vet
vet:
	@$(GO) vet ./...

.PHONY: mock
mock:
	@mockery

.PHONY: mock-clean
mock-clean:
	@find . \( -name '*_mock_test.go' -o -name '*_mock.go' \) -not -path './vendor/*' -delete

GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')
GOCACHE ?= $(CURDIR)/.go-cache

vendor: go.mod go.sum
	@$(GO) mod vendor

bin/litdoc: vendor $(GO_FILES)
	@GOCACHE=$(GOCACHE) $(GO) build -mod=vendor -o bin/litdoc .

.PHONY: build
build: bin/litdoc

.PHONY: test
test: vendor
	@GOCACHE=$(GOCACHE) $(GO) test -mod=vendor ./... --count=1

.PHONY: vendor-clean
vendor-clean:
	@rm -rf vendor/

.PHONY: clean
clean: mock-clean
	@rm -rf bin/
