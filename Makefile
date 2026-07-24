.PHONY: verify
verify:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || \
		{ echo "Go files need formatting:"; gofmt -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1; }
	go mod tidy -diff
	go vet ./...
	go test ./...
	buf lint
	$(MAKE) breaking

.PHONY: breaking
breaking:
	buf breaking . --against '.git#tag=v0.0.1'

.PHONY: generate
generate:
	buf generate
	buf format -w
	buf lint
	buf push
