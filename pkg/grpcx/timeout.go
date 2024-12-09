package grpcx

import (
	"time"
	"google.golang.org/grpc"
	"context"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

// TimeoutStreamInterceptor returns a new unary server interceptor for timeout.
func TimeoutStreamInterceptor(timeout time.Duration) grpc.StreamServerInterceptor {
	if timeout <= 0 {
		return func(
			srv any,
			ss grpc.ServerStream,
			_ *grpc.StreamServerInfo,
			handler grpc.StreamHandler,
		) error {
			return handler(srv, ss)
		}
	}

	return func(
		srv any,
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		ctx, cancel := context.WithTimeout(ss.Context(), timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- handler(srv, contextedStream{ctx, ss})
		}()

		select {
		case <-ctx.Done():
			return status.Error(codes.DeadlineExceeded, "timed out")
		case err = <-done:
			return err
		}
	}
}

type contextedStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (s contextedStream) Context() context.Context { return s.ctx }
