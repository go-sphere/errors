.PHONY: generate
generate:
	buf generate
	buf format -w
	buf lint
	buf push