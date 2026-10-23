# Static analysis tools are installed into .bin/ on first use, pinned to the versions below.
STATICCHECK_VERSION ?= v0.8.2
STATICCHECK_BIN := .bin/staticcheck
GOVULNCHECK_VERSION ?= v1.9.0
GOVULNCHECK_BIN := .bin/govulncheck
BIOME_VERSION ?= 2.5.14
HTMLHINT_VERSION ?= 1.1.4
# The templates are fragments with Go template conditionals, so only the rules that hold for them are used
HTMLHINT_RULES := attr-no-duplication,src-not-empty,attr-value-double-quotes

# The binary carries its platform in the name, so .bin/ can be shared between a host and a container
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
	BIOME_OS := darwin
else
	BIOME_OS := linux
endif
ifneq (,$(filter arm64 aarch64,$(UNAME_M)))
	BIOME_ARCH := arm64
else
	BIOME_ARCH := x64
endif
BIOME_BIN := .bin/biome-$(BIOME_OS)-$(BIOME_ARCH)

.PHONY: help build test test-e2e run lint lint-go lint-fmt lint-vet lint-staticcheck lint-govuln lint-html lint-js

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the logbogen binary
	go build -o logbogen .

test: ## Run all tests
	go test ./...

test-e2e: ## Run the browser tests in Chrome, they need network access, Google Chrome and Node
	go test -tags e2e -count=1 -run '^TestE2E' .

run: ## Run the app
	go run .

lint: lint-go lint-govuln lint-html lint-js ## Run all linters: Go, vulnerabilities, HTML and JS

lint-go: lint-fmt lint-vet lint-staticcheck ## Run gofmt, go vet and staticcheck

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

lint-govuln: $(GOVULNCHECK_BIN) ## Fail if the code calls a known vulnerable function
	$(GOVULNCHECK_BIN) ./...

$(GOVULNCHECK_BIN):
	GOBIN="$(CURDIR)/.bin" go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

lint-html: ## Check the HTML templates with htmlhint, needs Node
	npx --yes htmlhint@$(HTMLHINT_VERSION) --rules $(HTMLHINT_RULES) "views/**/*.html"

lint-js: $(BIOME_BIN) ## Lint the JavaScript and CSS files with Biome (config: biome.json)
	$(BIOME_BIN) lint .

$(BIOME_BIN):
	@mkdir -p .bin
	curl -sfL -o $(BIOME_BIN) \
		"https://github.com/biomejs/biome/releases/download/%40biomejs/biome%40$(BIOME_VERSION)/biome-$(BIOME_OS)-$(BIOME_ARCH)"
	chmod +x $(BIOME_BIN)
