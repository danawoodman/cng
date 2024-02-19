.PHONY: dev build install test watch-test

dev:
	@cng -ik '**/*.go' -- make build

# @cng -ik '**/*.go' -- go run ./cmd/cng $(ARGS)

test:
	@go test -v ./...

watch-test:
	@cng -ik '**/*.go' -- make install test

build:
	@CGO_ENABLED=0 go build -a -gcflags=all="-l -B" -ldflags="-s -w" -o bin/cng ./cmd/cng
	@echo "🎉 cng built to bin/cng"

install:
	@go install ./cmd/cng
	@echo "🎉 cng installed to: $(shell which cng)"

watch-install:
	@cng -ik '**/*.go' -- make install

.DEFAULT_GOAL := dev