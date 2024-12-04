// Package grpcx provides helper types and functions to work with gRPC.
//nolint:revive // ok to have no godoc on obvious methods
package grpcx

import (
	"log/slog"
	"os"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// LogInterceptor logs gRPC calls.
func LogInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	resp, err = handler(ctx, req)
	p, _ := peer.FromContext(ctx)
	slog.DebugContext(ctx, "gRPC call",
		slog.String("method", info.FullMethod),
		slog.Any("error", err),
		slog.Group("peer",
			slog.String("ip", p.Addr.String()),
			slog.String("local_ip", p.LocalAddr.String()),
			slog.Any("auth", p.AuthInfo != nil),
		))
	return resp, err
}

// Logger implements grpclog.LoggerV2 to log to slog.
type Logger struct{}

func (Logger) Info(args ...any)                    { slog.Info("grpclog", args...) }
func (Logger) Infoln(args ...any)                  { slog.Info("grpclog", args...) }
func (Logger) Warning(args ...any)                 { slog.Warn("grpclog", args...) }
func (Logger) Warningln(args ...any)               { slog.Warn("grpclog", args...) }
func (Logger) Infof(format string, args ...any)    { slog.Info("grpclog | "+format, args...) }
func (Logger) Warningf(format string, args ...any) { slog.Warn("grpclog | "+format, args...) }
func (Logger) Error(args ...any)                   { slog.Error("grpclog", args...) }
func (Logger) Errorln(args ...any)                 { slog.Error("grpclog", args...) }
func (Logger) Errorf(format string, args ...any)   { slog.Error("grpclog | "+format, args...) }
func (Logger) Fatal(args ...any)                   { slog.Error("grpclog | fatal", args...); os.Exit(1) }
func (Logger) Fatalln(args ...any)                 { slog.Error("grpclog | fatal", args...); os.Exit(1) }
func (Logger) Fatalf(format string, args ...any)   { slog.Error("grpclog | "+format, args...); os.Exit(1) }

func (Logger) V(l int) bool {
	lvl := slog.Level(-1)
	switch l {
	case 0:
		lvl = slog.LevelInfo
	case 1:
		lvl = slog.LevelWarn
	case 2, 3:
		lvl = slog.LevelError
	}

	return slog.Default().Enabled(context.Background(), lvl)
}
