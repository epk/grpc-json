## gRPC JSON

A PoC that demonstrates how to provide JSON compatibility for gRPC services without using gRPC-Gateway.

### How to run

Pre-requisites:
- grcpurl
```bash
brew install grpcurl
```
- curl
```bash
brew install curl
```

1. Run the gRPC server
```bash
go run .
```

2. gRPC client using grpcurl
```bash
# Unary
grpcurl -plaintext -d '{}' 0.0.0.0:50051 grpc.health.v1.Health/Check
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
