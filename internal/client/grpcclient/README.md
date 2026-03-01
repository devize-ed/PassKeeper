# Package grpcclient

gRPC client for the PassKeeper service.

## Auth interceptor

Injects a bearer token from `TokenStore` on all RPCs except Login and Register. If `TokenStore` is `nil`, no token is sent.
