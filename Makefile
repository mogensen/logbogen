.PHONY: help docker-build test-coverage

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

docker-build: ## build docker image
	docker build -t logbogen:latest .

docker-up: ## build-compose up
	docker-compose up -d

docker-down: ## build-compose up
	docker-compose down

clean-migrate: ## new test database
	buffalo pop drop
	buffalo pop create
	buffalo pop migrate up
	buffalo task demo:seed

test-coverage: ## Generate test coverage report
	mkdir -p tmp
	go test ./... --coverprofile tmp/outfile
	go tool cover -html=tmp/outfile

report-card: ## Generate static analysis report
	goreportcard-cli -v
