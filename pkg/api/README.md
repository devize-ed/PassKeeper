# Package api

Generated gRPC/protobuf types and service definitions for PassKeeper.

## Overview

- **PasskeeperAuthService**: Register, Login
- **PasskeeperItemService**: CreateItem, UpdateItem, DeleteItem, GetItem, ListItems

## Generating

```bash
protoc --go_out=. --go-grpc_out=. pkg/api/passkeper.proto
```
