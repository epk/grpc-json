## gRPC JSON

A proof of concept demonstrating how to have JSON REST and gRPC-web compability in a single gRPC server.

### How to run

Pre-requisites:
- grcpurl
- curl

## Demo

1. Run the gRPC server
```bash
go run .
```

2. gRPC client using grpcurl
```bash
# Unary
grpcurl -plaintext -d '{}' localhost:50051 grpc.health.v1.Health/Check
# Server streaming
grpcurl -plaintext -d '{}' localhost:50051 grpc.health.v1.Health/Watch
```

3. JSON client using curl
```bash
# Unary
curl -X POST -H "Content-Type: application/json" -d '{}' http://localhost:50051/grpc.health.v1.Health/Check
# Server streaming
curl -X POST -H "Content-Type: application/json" -d '{}' http://localhost:50051/grpc.health.v1.Health/Watch
```
