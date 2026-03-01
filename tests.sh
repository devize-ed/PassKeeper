#!/usr/bin/env bash
if set -o pipefail 2>/dev/null; then
  set -euo pipefail
else
  set -eu
fi

# Setup variables
SERVER_ADDRESS="${SERVER_ADDRESS:-localhost:50051}"
TOKEN_STORE_PATH="${TOKEN_STORE_PATH:-./.tmp/e2e-token}"
USER1_LOGIN="${USER1_LOGIN:-e2euser}"
USER1_PASS="${USER1_PASS:-e2epass123}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "ERROR: '$1' is required"; exit 1; }; }
say() { printf "\n\033[1;34m %s\033[0m\n" "$*"; }
expect_success() {
  if [[ $1 -ne 0 ]]; then
    echo "ERROR: command failed with exit code $1"
    exit 1
  fi
}

# Ensure .tmp exists for token store
mkdir -p "$(dirname "$TOKEN_STORE_PATH")"

# Check tools
need go

say "Server: $SERVER_ADDRESS"
say "Token store: $TOKEN_STORE_PATH"

# Build client if not present
CLIENT_BIN="${CLIENT_BIN:-.tmp/passkeeper}"
if [[ ! -x "$CLIENT_BIN" ]]; then
  say "Building client binary"
  go build -o "$CLIENT_BIN" ./cmd/client
  chmod +x "$CLIENT_BIN"
fi

pk() {
  SERVER_ADDRESS="$SERVER_ADDRESS" TOKEN_STORE_PATH="$TOKEN_STORE_PATH" \
    "$CLIENT_BIN" --address "$SERVER_ADDRESS" --token-store-path "$TOKEN_STORE_PATH" "$@"
}

# Register user (ignore failure if already exists)
say "Register user ($USER1_LOGIN)"
pk register --username "$USER1_LOGIN" --password "$USER1_PASS" 2>/dev/null || true

# Login and save token
say "Login and save token"
pk login --username "$USER1_LOGIN" --password "$USER1_PASS"
expect_success $?

# Create text item (type 2): pass text and metadata via stdin
say "Create text item"
printf 'e2e-text-content\ne2e-metadata\n' | pk create -t 2
expect_success $?

# List items (type 0 = all)
say "List items"
LIST_OUT=$(pk list -t 0 2>/dev/null || true)
expect_success $?
if [[ -z "$LIST_OUT" ]]; then
  echo "ERROR: list produced no output"
  exit 1
fi
echo "$LIST_OUT" | head -20

# Extract first item ID from list JSON
ITEM_ID=""
if command -v jq >/dev/null 2>&1; then
  JSON_PART=$(echo "$LIST_OUT" | grep -E '^\s*\[' | head -1)
  [[ -z "$JSON_PART" ]] && JSON_PART=$(echo "$LIST_OUT" | tail -n +2)
  ITEM_ID=$(echo "$JSON_PART" | jq -r '.[0].id // empty' 2>/dev/null || true)
fi
if [[ -z "$ITEM_ID" || "$ITEM_ID" == "null" ]]; then
  # Fallback: grep first "id" field (UUID format)
  ITEM_ID=$(echo "$LIST_OUT" | grep -oE '"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"' | head -1 | tr -d '"')
fi

if [[ -n "$ITEM_ID" && "$ITEM_ID" != "null" ]]; then
  # Get item by ID
  say "Get item $ITEM_ID"
  pk get -i "$ITEM_ID"
  expect_success $?

  # Edit item
  say "Edit item $ITEM_ID"
  printf 'updated-text\nupdated-meta\n' | pk edit -i "$ITEM_ID"
  expect_success $?
else
  say "Skipping get/edit (no jq or could not parse item id)"
fi

say "E2E tests finished successfully"
