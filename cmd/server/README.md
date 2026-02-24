# cmd/server

Entry point for the PassKeeper gRPC server.

Run via `go run ./cmd/server` with `-d` (DSN) and `-k` (auth secret) or equivalent environment variables. Gracefully shuts down on SIGINT/SIGTERM.
