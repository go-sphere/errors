.PHONY: generate
generate:
	buf generate
	buf format -w
	buf lint
	go fmt ./...
	golangci-lint fmt --no-config --enable gofmt,goimports
	golangci-lint run --no-config --fix