APP_NAME := para-nbody
CMD_DIR := ./cmd
BIN_DIR := ./bin
BIN := $(BIN_DIR)/$(APP_NAME)

DB_PATH ?= ./dataset/para-nbody-v2.store
STORE ?= ttl

.PHONY: help build run run-sqlite run-ttl test clean reset-db tidy

help:
	@echo "Targets:"
	@echo "  make build       Build $(CMD_DIR) -> $(BIN)"
	@echo "  make run         Build and run with STORE=$(STORE)"
	@echo "  make run-ttl     Build and run with TTL store"
	@echo "  make run-sqlite  Build and run with SQLite store"
	@echo "  make test        Run go test ./..."
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

run: build
	PNBODY_STORE=$(STORE) PNBODY_DB=$(DB_PATH) $(BIN)

run-ttl: build
	PNBODY_STORE=ttl $(BIN)

run-sqlite: build
	PNBODY_STORE=sqlite PNBODY_DB=$(DB_PATH) $(BIN)

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

reset-db:
	rm -f $(DB_PATH) $(DB_PATH)-wal $(DB_PATH)-shm