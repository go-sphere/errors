GO ?= go
GOLANGCI_LINT ?= golangci-lint

DIRECT_DEPS_TEMPLATE := {{if and (not .Main) (not .Indirect) (not .Replace)}}{{.Path}}{{end}}

.DEFAULT_GOAL := check

.PHONY: deps-update tidy fmt test lint check verify breaking generate

deps-update:
	@deps="$$(GOWORK=off $(GO) list -m -f '$(DIRECT_DEPS_TEMPLATE)' all)"; \
	if [ -n "$$deps" ]; then GOWORK=off $(GO) get -u $$deps; fi
	GOWORK=off $(GO) mod tidy

tidy:
	GOWORK=off $(GO) mod tidy

fmt:
	$(GO) fmt ./...
	$(GOLANGCI_LINT) fmt --no-config --enable gofmt --enable goimports
	buf format -w

test:
	$(GO) test ./...

lint:
	$(GOLANGCI_LINT) fmt --no-config --enable gofmt --enable goimports --diff
	$(GO) vet ./...
	$(GOLANGCI_LINT) run --no-config
	buf lint

check:
	GOWORK=off $(GO) mod tidy -diff
	$(MAKE) lint
	$(MAKE) test

verify: check breaking

breaking:
	buf breaking . --against '.git#tag=v0.0.1'

generate:
	buf generate
	buf format -w
	buf lint
	buf push
