package main

import (
	"context"
	"errors"
	"io"
	"iter"

	"google.golang.org/grpc"
	healthcheck "google.golang.org/grpc/health/grpc_health_v1"
)

// based on https://www.bwplotka.dev/2025/go-grpc-inprocess-iter
func newServerAsClient(srv healthcheck.HealthServer) healthcheck.HealthClient {
	return &serverAsClient{srv: srv}
}

type serverAsClient struct {
	srv healthcheck.HealthServer
}

func (s *serverAsClient) Check(ctx context.Context, in *healthcheck.HealthCheckRequest, opts ...grpc.CallOption) (*healthcheck.HealthCheckResponse, error) {
	return s.srv.Check(ctx, in)
}

func (s *serverAsClient) Watch(ctx context.Context, in *healthcheck.HealthCheckRequest, opts ...grpc.CallOption) (healthcheck.Health_WatchClient, error) {
	y := &yielder[healthcheck.HealthCheckResponse]{ctx: ctx}

	// Pull from iter.Seq2[*ListResponse, error].
	y.recv, y.stop = iter.Pull2(func(yield func(*healthcheck.HealthCheckResponse, error) bool) {
		y.send = yield
		if err := s.srv.Watch(in, y); err != nil {
			yield(nil, err)
			return
		}
	})

	return y, nil
}

type yielder[T any] struct {
	grpc.ServerStreamingClient[T]
	grpc.ServerStreamingServer[T]

	ctx context.Context

	send func(*T, error) bool
	recv func() (*T, error, bool)
	stop func()
}

func (y *yielder[T]) Context() context.Context { return y.ctx }

func (y *yielder[T]) Send(resp *T) error {
	if !y.send(resp, nil) {
		return errors.New("iterator stopped receiving")
	}
	return nil
}

func (y *yielder[T]) Recv() (*T, error) {
	r, err, ok := y.recv()
	if err != nil {
		y.stop()
		return nil, err
	}
	if !ok {
		return nil, io.EOF
	}
	return r, nil
}

func (y *yielder[T]) SendMsg(m any) error {
	return y.Send(m.(*T))
}

func (y *yielder[T]) RecvMsg(m any) error {
	r, err := y.Recv()
	if err != nil {
		return err
	}

	*m.(*T) = *r
	return nil
}
