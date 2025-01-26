package main

import (
	"context"
	"time"

	healthcheck "google.golang.org/grpc/health/grpc_health_v1"
)

// gRPC implementation of healthcheck.HealthServer
var _ healthcheck.HealthServer = &healtcheckServer{}

type healtcheckServer struct{}

func (s *healtcheckServer) Check(ctx context.Context, in *healthcheck.HealthCheckRequest) (*healthcheck.HealthCheckResponse, error) {
	return &healthcheck.HealthCheckResponse{
		Status: healthcheck.HealthCheckResponse_SERVING,
	}, nil
}

func (s *healtcheckServer) Watch(in *healthcheck.HealthCheckRequest, srv healthcheck.Health_WatchServer) error {
	for i := 0; i < 5; i++ {
		resp := &healthcheck.HealthCheckResponse{
			Status: healthcheck.HealthCheckResponse_SERVING,
		}
		if err := srv.Send(resp); err != nil {
			return err
		}
		select {
		case <-srv.Context().Done():
			return srv.Context().Err()
			// Simulate periodic updates with a delay
		case <-time.After(2 * time.Second):
		}
	}
	return nil
}
