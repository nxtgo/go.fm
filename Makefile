ifneq (,$(wildcard .env))
    include .env
    export
endif

module := go.fm
bin := gofm
pkg := ./cmd/bot

GO ?= go
gofmt := gofmt
goimports := goimports
golangci_lint := golangci-lint

.PHONY: all
all: tidy fmt vet lint test build

.PHONY: run
run:
	@echo ">> running $(bin)..."
	$(GO) run $(pkg)

.PHONY: build
build:
	@echo ">> building $(bin)..."
	GOOS=$(shell $(GO) env GOOS) GOARCH=$(shell $(GO) env GOARCH) \
		$(GO) build -o bin/$(bin) $(pkg)

.PHONY: install
install:
	@echo ">> installing $(bin)..."
	$(GO) install $(pkg)

.PHONY: clean
clean:
	@echo ">> cleaning build artifacts..."
	rm -rf bin coverage.out coverage.html

.PHONY: fmt
fmt:
	@echo ">> formatting source code..."
	$(gofmt) -s -w .

.PHONY: fmt-check
fmt-check:
	@echo ">> checking formatting..."
	@$(gofmt) -l .

.PHONY: tidy
tidy:
	@echo ">> tidying modules..."
	$(GO) mod tidy

.PHONY: vet
vet:
	@echo ">> running go vet..."
	$(GO) vet ./...

.PHONY: lint
lint:
	@echo ">> running linter..."

.PHONY: test
test:
	@echo ">> running tests..."
	$(GO) test -v -race -cover ./...

.PHONY: cover
cover:
	@echo ">> generating coverage report..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo ">> open coverage.html in your browser"

.PHONY: update
update:
	@echo ">> updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: outdated
outdated:
	@echo ">> checking outdated dependencies..."
	@$(GO) list -u -m -json all | $(GO) run golang.org/x/exp/cmd/modoutdated
