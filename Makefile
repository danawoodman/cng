.PHONY: dev build install

dev:
	@gochange -i -k '**/*.go' -- go run ./cmd/gochange $(ARGS)

build:
	@go build -o bin/gochange ./cmd/gochange
	@echo "gochange built to bin/gochange"

install:
	@go install ./cmd/gochange
	@echo "gochange installed to $(shell which gochange)"

.DEFAULT_GOAL := dev