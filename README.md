## PassKeeper (Password Manager)

A client-server password manager built with Go and gRPC. Register, authenticate, and securely store credentials, text notes, binary data, and card details.

## Prerequisites

### System Requirements
- **Go**: Version 1.24 or higher
- **Docker & Docker Compose**: For running PostgreSQL database
- **Make**: For running build and test commands

### Devtools
- **Mockery**: For generating mocks (see `cfg/.mockery.yaml`)
- **protoc**: For regenerating gRPC/protobuf code (optional)

## Installation & Setup

**Generate mocks**:
```bash
make mock-gen
```

## Running the Service

### 1. Start Database
Start PostgreSQL using Docker Compose:
```bash
make up
```

### 2. Run the Server

```bash
make run
```

### 3. Stop the Service
Stop background server:
```bash
make stop
```

Stop and clean up containers:
```bash
make down
```

## Configuration

The service can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDRESS` | `localhost:50051` | gRPC server address and port |
| `DATABASE_DSN` | `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable` | PostgreSQL connection string |
| `AUTH_SECRET` | `` | JWT signing secret (required for server) |
| `TOKEN_STORE_PATH` | `~/.passkeeper/token` | Client token file path |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

Custom configuration:
```bash
SERVER_ADDRESS=localhost:50052 DATABASE_DSN="postgres://user:pass@localhost:5432/mydb" AUTH_SECRET="my-secret" make run
```

## Testing

**Run all tests**:
```bash
make tests
```

**Run unit/integration tests**:
```bash
go test ./... -v --tags=integration_tests
```

**Run end-to-end tests**:
```bash
make e2e
```
## CLI Usage Guide

The PassKeeper client is a command-line tool for managing stored secrets. Parameters can be provided via **flags** or entered in **interactive mode** when a flag is omitted.

> **Note:** Ensure the server is running (`make run`) and the database is up (`make up`) before using the client. Build the client with `make build-client-linux` (or `build-client-mac` / `build-client-windows`) for your platform; binaries go to `cmd/client/`. For quick local runs use `go run ./cmd/client`.

### Global Flags

Available for all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--address` | - | `localhost:50051` | gRPC server address |
| `--token-store-path` | - | `~/.passkeeper/token` | Path to store the auth token |


### Commands

#### `register` — Create a new account

Creates a new user account. You must register before logging in.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--username` | `-u` | No* | Username |
| `--password` | `-p` | No* | Password |

\* If omitted, prompts for input in interactive mode.

```bash
# With flags
./passkeeper register -u myuser -p mypassword

# Interactive (will prompt for username and password)
./passkeeper register
```

---

#### `login` — Authenticate and save token

Logs in and stores the JWT token locally. All item commands require a valid token.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--username` | `-u` | No* | Username |
| `--password` | `-p` | No* | Password |

\* If omitted, prompts in interactive mode.

```bash
./passkeeper login -u myuser -p mypassword
# or: .tmp/passkeeper login
```

---

#### `create` — Add a new item

Creates a new item of the given type. Item data can be entered interactively or piped via stdin (for text/binary).

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--item-type` | `-t` | No* | Item type (1–4, see table below) |

\* If omitted or invalid, prompts for type.

**Item types:**

| Type | Value | Description | Stdin format (for piping) |
|------|-------|-------------|----------------------------|
| Credential | `1` | Login + password + metadata | Interactive only (password hidden) |
| Text | `2` | Text + metadata | `line1\nline2` (text, then metadata) |
| Binary | `3` | Binary data + metadata | `line1\nline2` (data, then metadata) |
| Card | `4` | Card name, number, exp, CVV | Interactive only (sensitive fields) |

```bash
# Text item — pipe content (line1=text, line2=metadata)
printf 'my secret note\noptional metadata\n' | ./passkeeper create -t 2
# or: echo -e "my note\nmy meta" | ./passkeeper create -t 2

# Binary item — pipe content
printf 'base64-or-raw-data\nmetadata\n' | ./passkeeper create -t 3

# Credential or Card — interactive (prompts for each field)
./passkeeper create -t 1
./passkeeper create -t 4
```

---

#### `list` — List items

Lists items, optionally filtered by type.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--item-type` | `-t` | `0` | Filter: 0=all, 1=credential, 2=text, 3=binary, 4=card |

```bash
# List all items
./passkeeper list
./passkeeper list -t 0

# List only text items
./passkeeper list -t 2

# List only credentials
./passkeeper list -t 1
```

---

#### `get` — Retrieve an item by ID

Fetches and displays a single item by its ID.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--item-id` | `-i` | No* | Item UUID (from `list` output) |

\* If omitted, prompts in interactive mode.

```bash
./passkeeper get -i a1b2c3d4-e5f6-7890-abcd-ef1234567890
# or: ./passkeeper get
```

---

#### `edit` — Update an existing item

Updates an item by ID. For text/binary items, new content can be piped via stdin.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--item-id` | `-i` | No* | Item UUID to edit |

\* If omitted, prompts in interactive mode.

```bash
# Text item — pipe new content
printf 'updated text\nupdated metadata\n' | ./passkeeper edit -i <item-id>

# Interactive (all item types)
./passkeeper edit -i <item-id>
```

---

#### `version` — Show version

Displays version and build date.

```bash
.tm/passkeeper version
```

---

### Example workflow

```bash
# Start server and DB first
make up
make run   # in another terminal

# Build client (e.g. for Linux)
make build-client-linux

# Account setup
./passkeeper register -u alice -p secret123
./passkeeper login -u alice -p secret123

# Store a text note
echo -e "My WiFi password: xyz123\nHome network" | ./passkeeper create -t 2

# List and get
./passkeeper list -t 0
./passkeeper get -i <id-from-list>

# Update
echo -e "Updated password\nHome network" | ./passkeeper edit -i <item-id>
```

## Building

Only four build commands (server → `cmd/server/`, clients → `cmd/client/`):

```bash
make build-server        # Server (host)
make build-client-linux  # Client for Linux (amd64)
make build-client-mac   # Client for macOS (Intel + Apple Silicon)
make build-client-windows # Client for Windows (amd64)
```
