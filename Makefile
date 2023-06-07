.DEFAULT_GOAL := all

ROOT_PATH := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
COVERAGE_PATH := $(ROOT_PATH).coverage/

EXECUTABLE=dbdiff

clean:
	@rm -rf $(BUILD_PATH)

dirs:
	@mkdir -p $(COVERAGE_PATH)

test: dirs
	@go test -v -coverpkg=./... ./... -coverprofile $(COVERAGE_PATH)cp.out
	@go tool cover -func=$(COVERAGE_PATH)cp.out -o $(COVERAGE_PATH)coverage.txt
	@go tool cover -html=$(COVERAGE_PATH)cp.out -o $(COVERAGE_PATH)coverage.html

lint:
	@golangci-lint run

build: dirs
	@goreleaser release --snapshot --skip-publish --clean

.PHONY: all
all: clean lint test build