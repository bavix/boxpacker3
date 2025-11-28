.PHONY: *

test:
	go test -tags mock -race -cover ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2 run --color always ${args}

lint-fix:
	make lint args=--fix
