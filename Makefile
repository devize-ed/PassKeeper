SHELL := /usr/bin/env bash
.SHELLFLAGS := -e -o pipefail -c
PROJECT_NAME := passkeeper

RUN_ADDRESS ?= localhost:50051

DB_DSN_HOST ?= localhost
DB_DSN      ?= postgres://postgres:postgres@$(DB_DSN_HOST):5432/postgres?sslmode=disable

AUTH_SECRET       ?= e2e-secret-key
TOKEN_STORE_PATH  ?= $(shell pwd)/.tmp/e2e-token

GO          ?= go
PKG         ?= ./...
GOFLAGS     ?=
TEST_FLAGS  ?= -count=1 -v

DC := docker compose -p $(PROJECT_NAME)

PIDFILE := .tmp/server.pid

MOCKERY := $(shell go env GOPATH)/bin/mockery

.PHONY: up down logs build-server build-client-linux build-client-mac build-client-windows \
        run run-bg wait-db wait-ready stop status test e2e e2e-keep tests mock-gen

## up: start Postgres
up:
	@echo " Starting Postgres..."
	$(DC) up -d db

## down: stop everything and remove volumes
down:
	$(DC) down -v

## logs: tail docker compose logs
logs:
	$(DC) logs -f --tail=200

## wait-db: wait until Postgres is ready (and accept connections from host)
wait-db:
	@echo " Waiting for Postgres..."
	@for i in $$(seq 1 60); do \
		$(DC) exec -T db pg_isready -U postgres -d postgres -h localhost -p 5432 >/dev/null 2>&1 && { echo "Postgres is ready"; sleep 2; exit 0; }; \
		sleep 1; \
	done; \
	echo "ERROR: Postgres is not ready"; exit 1

## build-server: build server (host platform) in cmd/server
build-server:
	@mkdir -p cmd/server
	@echo " Building server..."
	@CGO_ENABLED=0 $(GO) build -o cmd/server/passkeeper-server ./cmd/server
	@echo " Done. cmd/server/passkeeper-server"

## build-client-linux: build client for Linux (amd64) in cmd/client
build-client-linux:
	@mkdir -p cmd/client
	@echo " Building client for Linux..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o cmd/client/passkeeper ./cmd/client
	@echo " Done. cmd/client/passkeeper"

## build-client-mac: build client for macOS (Intel + Apple Silicon) in cmd/client
build-client-mac:
	@mkdir -p cmd/client
	@echo " Building client for macOS (Intel)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -o cmd/client/passkeeper ./cmd/client
	@echo " Building client for macOS (Apple Silicon)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -o cmd/client/passkeeper ./cmd/client
	@echo " Done. cmd/client/passkeeper"

## build-client-windows: build client for Windows (amd64) in cmd/client
build-client-windows:
	@mkdir -p cmd/client
	@echo " Building client for Windows..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -o cmd/client/passkeeper.exe ./cmd/client
	@echo " Done. cmd/client/passkeeper.exe"

## run: run server in foreground
run:
	SERVER_ADDRESS="$(RUN_ADDRESS)" DATABASE_DSN="$(DB_DSN)" AUTH_SECRET="$(AUTH_SECRET)" \
	$(GO) run ./cmd/server

## run-bg: run server in background
run-bg:
	mkdir -p .tmp
	@echo " Starting server on $(RUN_ADDRESS) with DB=$(DB_DSN)"
	@set -e; \
	SERVER_ADDRESS="$(RUN_ADDRESS)" DATABASE_DSN="$(DB_DSN)" AUTH_SECRET="$(AUTH_SECRET)" \
	$(GO) run ./cmd/server > .tmp/server.log 2>&1 & echo $$! > $(PIDFILE); \
	sleep 0.5; \
	if ! kill -0 $$(cat $(PIDFILE)) 2>/dev/null; then \
		echo "ERROR: server exited immediately"; \
		tail -n 200 .tmp/server.log || true; \
		exit 1; \
	fi; \
	echo "PID=$$(cat $(PIDFILE))"

## wait-ready: wait until server is listening
wait-ready:
	@PORT=$$(echo $(RUN_ADDRESS) | awk -F: '{print $$NF}'); \
	echo " Waiting for server on port $$PORT ..."; \
	for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do \
		(nc -z 127.0.0.1 $$PORT 2>/dev/null) && { echo "Server is up"; exit 0; }; \
		(lsof -i:$$PORT -sTCP:LISTEN 2>/dev/null | grep -q .) && { echo "Server is up"; exit 0; }; \
		(ss -tlnp 2>/dev/null | grep -q ":$$PORT ") && { echo "Server is up"; exit 0; }; \
		if [ -f $(PIDFILE) ] && ! kill -0 $$(cat $(PIDFILE)) 2>/dev/null; then \
			echo "ERROR: server process died early"; \
			tail -n 200 .tmp/server.log || true; \
			exit 1; \
		fi; \
		sleep 0.5; \
	done; \
	echo "ERROR: server did not start in time"; \
	tail -n 200 .tmp/server.log || true; \
	exit 1

## stop: stop background server
stop:
	@PORT=$$(echo $(RUN_ADDRESS) | awk -F: '{print $$NF}'); \
	if [ -f $(PIDFILE) ]; then \
		echo " Stopping server PID $$(cat $(PIDFILE))"; \
		kill -TERM $$(cat $(PIDFILE)) 2>/dev/null || true; \
		for i in 1 2 3 4 5; do \
			kill -0 $$(cat $(PIDFILE)) 2>/dev/null || break; \
			sleep 1; \
		done; \
		kill -0 $$(cat $(PIDFILE)) 2>/dev/null && kill -KILL $$(cat $(PIDFILE)) || true; \
		rm -f $(PIDFILE); \
	fi; \
	if lsof -ti:$$PORT -sTCP:LISTEN >/dev/null 2>/dev/null || ss -tlnp 2>/dev/null | grep -q ":$$PORT "; then \
		PIDS=$$(lsof -ti:$$PORT -sTCP:LISTEN 2>/dev/null || true); \
		[ -n "$$PIDS" ] && { echo " Stopping PIDs: $$PIDS"; kill -TERM $$PIDS 2>/dev/null || true; sleep 1; kill -KILL $$PIDS 2>/dev/null || true; }; \
	fi

status:
	@PORT=$$(echo $(RUN_ADDRESS) | awk -F: '{print $$NF}'); \
	echo "RUN_ADDRESS=$(RUN_ADDRESS)  PORT=$$PORT"; \
	if [ -f $(PIDFILE) ]; then echo "PIDFILE=$(PIDFILE): $$(cat $(PIDFILE))"; else echo "PIDFILE: (none)"; fi; \
	lsof -i:$$PORT -sTCP:LISTEN -nP 2>/dev/null || ss -tlnp 2>/dev/null | grep ":$$PORT " || echo "Port $$PORT is free"

# Tests
## test: run unit tests
test:
	@echo " Running unit tests"
	$(GO) test $(PKG) $(TEST_FLAGS) $(GOFLAGS)

## e2e: run end-to-end tests
e2e:
	@echo " Running e2e tests"
	@trap '$(MAKE) stop; $(DC) down -v' EXIT; \
	$(MAKE) stop; \
	$(MAKE) up; \
	$(MAKE) wait-db; \
	$(MAKE) run-bg; \
	$(MAKE) wait-ready; \
	SERVER_ADDRESS="$(RUN_ADDRESS)" TOKEN_STORE_PATH="$(TOKEN_STORE_PATH)" \
	bash -eu -o pipefail ./tests.sh

## e2e-keep: run e2e without tearing down containers
e2e-keep:
	@trap '$(MAKE) stop; $(DC) down' EXIT; \
	$(MAKE) stop; \
	$(MAKE) up; \
	$(MAKE) wait-db; \
	$(MAKE) run-bg; \
	$(MAKE) wait-ready; \
	SERVER_ADDRESS="$(RUN_ADDRESS)" TOKEN_STORE_PATH="$(TOKEN_STORE_PATH)" \
	bash -eu -o pipefail ./tests.sh

tests: test e2e
	@echo "All tests passed"

## mock-gen: generate mocks
mock-gen: mockery-install
	@echo " Generating mocks"
	$(MOCKERY) --config cfg/.mockery.yaml
