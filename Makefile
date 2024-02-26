#------------------------------------------------------------------------------
# TESTS
#------------------------------------------------------------------------------

.PHONY: test
test:
	@make -j test-unit test-e2e

.PHONY: test-e2e
test-e2e:
	@go test ./test/...

.PHONY: test-unit
test-unit:
	@go test ./internal/...

.PHONY: watch-test
watch-test:
	@make -j watch-unit-test watch-e2e-test

.PHONY: watch-unit-test
watch-unit-test:
	@cng -ik -e 'test' '**/*.go' -- make test-unit

.PHONY: watch-e2e-test
watch-e2e-test:
	@cng -ik '**/*.go' -- make install test-e2e

#------------------------------------------------------------------------------
# BUILDING
#------------------------------------------------------------------------------

.PHONY: build
build:
	@echo "🤖 building cng..."
	@CGO_ENABLED=0 go build -a -gcflags=all="-l -B" -ldflags="-s -w" -o ./dist/cng ./cmd/cng
	@echo "🎉 cng built to dist/cng"

.PHONY: install
install:
	@echo "🤖 installing cng..."
	@go install ./cmd/cng
	@echo "🎉 cng installed to: $(shell which cng)"

.PHONY: watch-install
watch-install:
	@cng -ik '**/*.go' -- make install

.DEFAULT_GOAL := watch-test