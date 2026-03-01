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

The service can be configured via **config files** (`cfg/client.yaml`, `cfg/server.yaml`), **environment variables**, and **command-line flags** (server only). See `cfg/example_client.yaml` and `cfg/example_server.yaml` for all options.

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDRESS` | `localhost:50051` | gRPC server address and port |
| `DATABASE_DSN` | (see Makefile) | PostgreSQL connection string |
| `AUTH_SECRET` | (required) | JWT signing secret (server) |
| `TOKEN_STORE_PATH` | `~/.passkeeper/token` | Client token file path |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| **Server TLS** | | |
| `TLS_ENABLED` | `false` | Enable TLS for gRPC |
| `TLS_CERT_FILE` | — | Path to PEM certificate (required if TLS enabled) |
| `TLS_KEY_FILE` | — | Path to PEM private key (required if TLS enabled) |
| **Client TLS** | | |
| `TLS_ENABLED` | `false` | Use TLS when connecting to server |
| `TLS_CERT_FILE` | — | Path to server cert or CA PEM (required if TLS enabled) |
| `TLS_SERVER_NAME` | — | Server hostname for verification (required if TLS enabled) |

Custom configuration:
```bash
SERVER_ADDRESS=localhost:50052 DATABASE_DSN="postgres://user:pass@localhost:5432/mydb" AUTH_SECRET="my-secret" make run
```

### TLS / self-signed certificates

gRPC can run over TLS with **CA-signed or self-signed** certificates.

**Server:** Set `tls_enabled: true` and provide `tls_cert_file` and `tls_key_file` (PEM paths). You can use a self-signed cert (e.g. `openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes`). Override via flags: `-t` (enable TLS), `-c` (cert file), `-b` (key file).

**Client:** Set `tls_enabled: true`, set `tls_cert_file` to the **server’s certificate** (or the CA that signed it), and set `tls_server_name` to the server hostname (e.g. `localhost`). The client trusts the server by using that cert file;

Example with config files: copy `cfg/example_server.yaml` to `cfg/server.yaml` and `cfg/example_client.yaml` to `cfg/client.yaml`, then set the TLS fields and paths as above.

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
