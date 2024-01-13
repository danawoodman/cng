.PHONY: dev build install

dev:
	@gochange -i -k '**/*.go' -- go run ./ $(ARGS)

build:
	@CGO_ENABLED=0 go build -a -gcflags=all="-l -B" -ldflags="-s -w" -o bin/gochange ./
	@echo "gochange built to bin/gochange"

install:
	@go install ./
	@echo "gochange installed to $(shell which gochange)"

.DEFAULT_GOAL := dev