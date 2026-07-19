APP_NAME := para-nbody
CMD_DIR := ./cmd
BIN_DIR := ./bin
BIN := $(BIN_DIR)/$(APP_NAME)

DB_PATH ?= ./dataset/para-nbody-v2.store
STORE ?= ttl

.PHONY: help build run run-sqlite run-ttl test test-v test-pkg clean reset-db tidy

help:
	@echo "Targets:"
	@echo "  make build       Build $(CMD_DIR) -> $(BIN)"
	@echo "  make run         Build and run with STORE=$(STORE)"
	@echo "  make run-ttl     Build and run with TTL store"
	@echo "  make run-sqlite  Build and run with SQLite store"
	@echo "  make test        Run go test ./..."
	@echo "  make test-v      Run go test -v ./... (Verbose tracking outputs)"
	@echo "  make test-pkg    Run specific package tests (e.g., make test-pkg PKG=./types)"
	@echo "  make tidy        Run go mod tidy"
	@echo "  make clean       Remove ./bin"
	@echo "  make reset-db    Remove SQLite DB/WAL/SHM files"
	@echo ""
	@echo "Variables:"
	@echo "  STORE=ttl|sqlite"
	@echo "  DB_PATH=./dataset/para-nbody-v2.store"

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD_DIR)

portal:
	@echo "🎨 Compiling Standalone Embedded Web Portal..."
	# Run templ compiler generation first before wrapping the binary
	templ generate
	go build -o bin/gson-portal ./cmd/portal/web-main.go

run: build
	PNBODY_STORE=$(STORE) PNBODY_DB=$(DB_PATH) $(BIN)

run-ttl: build
	PNBODY_STORE=ttl $(BIN)

run-sqlite: build
	PNBODY_STORE=sqlite PNBODY_DB=$(DB_PATH) $(BIN)

build-linux:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/verify_v3-linux-amd64 ./cmd/verify_v3/main.go

test:
	@echo "🧪 Running tests while filtering out scripts..."
	go test $$(go list ./... | grep -v '/scripts')

test-v:
	@echo "🧪 Running full test suite with verbose audit logs (excluding scripts)..."
	go test -v $$(go list ./... | grep -v '/scripts')
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "❌ Error: Please specify a package target. Example: make test-pkg PKG=./types"; \
		exit 1; \
	fi
	@echo "🔬 Auditing isolated package domain: $(PKG)"
	go test -v $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

reset-db:
	rm -f $(DB_PATH) $(DB_PATH)-wal $(DB_PATH)-shm

gem:
	find . -name "*.go" -not -path "*/.*" -exec sh -c 'for f; do echo "=== FILE: $f ==="; cat "$f"; echo "\n"; done' _ {} + > consolidated_code.txt

git:
	git add . && git commit -m "clean up v3" && git push