# Static analysis tools are installed into .bin/ on first use, pinned to the versions below.
STATICCHECK_VERSION ?= v0.8.1
STATICCHECK_BIN := .bin/staticcheck

.PHONY: help build test test-e2e run lint lint-fmt lint-vet lint-staticcheck

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the logbogen binary
	go build -o logbogen .

test: ## Run all tests
	go test ./...

test-e2e: ## Run the browser tests in Chrome, they need network access and Google Chrome
	go test -tags e2e -count=1 -run '^TestE2E' .

run: ## Run the app
	go run .

lint: lint-fmt lint-vet lint-staticcheck ## Run gofmt, go vet and staticcheck

lint-fmt: ## Fail if any file is not gofmt formatted
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files need gofmt:"; echo "$$unformatted"; exit 1; \
	fi

lint-vet: ## Run go vet, including the browser tests
	go vet -tags e2e ./...

lint-staticcheck: $(STATICCHECK_BIN) ## Run staticcheck
	$(STATICCHECK_BIN) -tags e2e ./...

$(STATICCHECK_BIN):
	GOBIN="$(CURDIR)/.bin" go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
