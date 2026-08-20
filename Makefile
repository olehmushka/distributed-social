
SHELL = /bin/bash
.SHELLFLAGS = -o pipefail -c

# https://github.com/golang/go/wiki/LoopvarExperiment
export GOEXPERIMENT := loopvar

.PHONY: help
help: ## Print info about all commands
	@echo "Commands:"
	@echo
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[01;32m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build all executables
	go build ./cmd/accounts
	go build ./cmd/admins
	go build ./cmd/search

.PHONY: all
all: build fmt lint test

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: test-short
test-short: ## Run tests, skipping slower integration tests
	go test -test.short ./...

.PHONY: coverage-html
coverage-html: ## Generate test coverage report and open in browser
	go test ./... -coverpkg=./... -coverprofile=test-coverage.out
	go tool cover -html=test-coverage.out

.PHONY: lint
lint: ## Verify code style and run static checks
	go vet -asmdecl -assign -atomic -bools -buildtag -cgocall -copylocks -httpresponse -loopclosure -lostcancel -nilfunc -printf -shift -stdmethods -structtag -tests -unmarshal -unreachable -unsafeptr -unusedresult ./...
	test -z $(gofmt -l ./...)

.PHONY: fmt
fmt: ## Run syntax re-formatting (modify in place)
	go fmt ./...

.PHONY: check
check: ## Compile everything, checking syntax (does not output binaries)
	go build ./...

.env:
	if [ ! -f ".env" ]; then cp example.dev.env .env; fi

.PHONY: up
up: ## Start nats + all three services' postgres instances for local dev
	docker compose up -d nats accounts-db admins-db search-db

.PHONY: down
down: ## Tear down the local dev dependencies (drops their data volumes)
	docker compose down -v

.PHONY: up-all
up-all: ## Build and run the full stack (services included) in containers
	docker compose up -d --build

.PHONY: logs
logs: ## Tail logs for the full stack
	docker compose logs -f

.PHONY: run-dev-accounts
run-dev-accounts: .env ## Runs accounts for local dev (needs `make up` first)
	ACCOUNTS_DB_DSN=postgres://postgres:postgres@localhost:5433/accounts?sslmode=disable \
	NATS_URL=nats://localhost:4222 \
	GOLOG_LOG_LEVEL=info \
	go run ./cmd/accounts

.PHONY: run-dev-admins
run-dev-admins: .env ## Runs admins for local dev (needs `make up` first)
	ADMINS_DB_DSN=postgres://postgres:postgres@localhost:5434/admins?sslmode=disable \
	NATS_URL=nats://localhost:4222 \
	GOLOG_LOG_LEVEL=info \
	go run ./cmd/admins

.PHONY: run-dev-search
run-dev-search: .env ## Runs search for local dev (needs `make up` first)
	SEARCH_DB_DSN=postgres://postgres:postgres@localhost:5435/search?sslmode=disable \
	NATS_URL=nats://localhost:4222 \
	GOLOG_LOG_LEVEL=info \
	go run ./cmd/search
