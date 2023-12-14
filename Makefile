
SHELL = /bin/bash
.SHELLFLAGS = -o pipefail -c

# base path for Lexicon document tree (for lexgen)
LEXDIR?=../atproto/lexicons

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

.PHONY: run-dev-accounts
run-dev-accounts: .env ## Runs accounts for local dev
	GOLOG_LOG_LEVEL=info go run ./cmd/accounts

.PHONY: build-accounts-image
build-accounts-image:
	docker build -t accounts -f cmd/accounts/Dockerfile .

.PHONY: run-accounts-image
run-accounts-image:
	docker run -p 9010:9010 accounts /accounts

.PHONY: run-dev-admins
run-dev-admins: .env ## Runs admins for local dev
	GOLOG_LOG_LEVEL=info go run ./cmd/admins

.PHONY: build-admins-image
build-admins-image:
	docker build -t admins -f cmd/admins/Dockerfile .

.PHONY: run-admins-image
run-admins-image:
	docker run -p 9011:9011 admins /admins

.PHONY: run-dev-search
run-dev-search: .env ## Runs search for local dev
	GOLOG_LOG_LEVEL=info go run ./cmd/search

.PHONY: build-search-image
build-search-image:
	docker build -t search -f cmd/search/Dockerfile .

.PHONY: run-search-image
run-search-image:
	docker run -p 9012:9012 search /search
