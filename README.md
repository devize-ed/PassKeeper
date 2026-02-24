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
> **Note:** If any of flag are not setted, it will be asked in the interactive mode.
#### `register` — Create a new account

Creates a new user account. You must register before logging in.

| Flag | Short | Description |
|------|-------|-------------|
| `--username` | `-u` | Username |
| `--password` | `-p` | Password |

\* If omitted, prompts for input in interactive mode.

```bash
./passkeeper register -u myuser -p mypassword
./passkeeper register
```

---

#### `login` — Authenticate and save token

Logs in and stores the JWT token locally. All item commands require a valid token.

| Flag | Short | Description |
|------|-------|-------------|
| `--username` | `-u` | Username |
| `--password` | `-p` | Password |

\* If omitted, prompts in interactive mode.

```bash
./passkeeper login -u myuser -p mypassword
./passkeeper login
```

---

#### `create` — Add a new item

Creates a new item of the given type. Item data can be entered interactively or piped via stdin (for text/binary).

| Flag | Short | Description |
|------|-------|-------------|
| `--item-type` | `-t` | Item type (1–4, see table below) |

```bash
./passkeeper create -t 1
./passkeeper create
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

| Flag | Short | Description |
|------|-------|-------------|
| `--item-id` | `-i` | Item UUID (from `list` output) |


```bash
./passkeeper get -i a1b2c3d4-e5f6-7890-abcd-ef1234567890
./passkeeper get
```

---

#### `edit` — Update an existing item

Updates an item by ID. For text/binary items, new content can be piped via stdin.

| Flag | Short | Description |
|------|-------|-------------|
| `--item-id` | `-i` | Item UUID to edit |

```bash
./passkeeper edit -i <item-id>
./passkeeper edit
```

---

#### `version` — Show version

Displays version and build date.

```bash
.tm/passkeeper version
```

## Building

Only four build commands (server → `cmd/server/`, clients → `cmd/client/`):

```bash
make build-server        # Server (host)
make build-client-linux  # Client for Linux (amd64)
make build-client-mac   # Client for macOS (Intel + Apple Silicon)
make build-client-windows # Client for Windows (amd64)
```
