# Static analysis tools are installed into .bin/ on first use, pinned to the versions below.
STATICCHECK_VERSION ?= v0.8.1
STATICCHECK_BIN := .bin/staticcheck

.PHONY: help build test run lint lint-fmt lint-vet lint-staticcheck

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the logbogen binary
	go build -o logbogen .

test: ## Run all tests
	go test ./...

run: ## Run the app
	go run .

lint: lint-fmt lint-vet lint-staticcheck ## Run gofmt, go vet and staticcheck

lint-fmt: ## Fail if any file is not gofmt formatted
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files need gofmt:"; echo "$$unformatted"; exit 1; \
	fi

lint-vet: ## Run go vet
	go vet ./...

lint-staticcheck: $(STATICCHECK_BIN) ## Run staticcheck
	$(STATICCHECK_BIN) ./...

$(STATICCHECK_BIN):
	GOBIN="$(CURDIR)/.bin" go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
