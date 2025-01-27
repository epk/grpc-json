package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	healthcheck "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}

	// Create a gRPC server
	s := grpc.NewServer()
	h := &healtcheckServer{}
	// Register health check service
	healthcheck.RegisterHealthServer(s, h)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	wrappedGrpc := grpcweb.WrapServer(s)
	// http mux to route based on content-type
	mux := http.NewServeMux()

	mux.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			// early return not POST
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			contentType := r.Header.Get("Content-Type")
			// handle gRPC and REST JSON requests
			switch contentType {
			case "application/grpc":
				s.ServeHTTP(w, r)
				return
			case "application/json":
				NewJSONHandler(newServerAsClient(h)).ServeHTTP(w, r)
				return
			}

			// handle gRPC-Web requests
			if wrappedGrpc.IsGrpcWebRequest(r) {
				wrappedGrpc.ServeHTTP(w, r)
				return
			}

			// default fall through
			http.Error(w, "unsupported content-	type", http.StatusUnsupportedMediaType)
		},
	)

	// wrap in h2c handler
	hh := h2c.NewHandler(mux, &http2.Server{})

	// Serve HTTP
	fmt.Print("Serving HTTP on 0.0.0.0:50051\n")
	if err := http.Serve(lis, hh); err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}
}
