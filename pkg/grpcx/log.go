// Package grpcx provides helper types and functions to work with gRPC.
//nolint:revive,govet // ok to have no godoc on obvious methods, ok for format without placeholders
package grpcx

import (
	"log/slog"
	"os"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"time"
)

// LogUnaryInterceptor logs gRPC calls.
func LogUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	now := time.Now()
	resp, err = handler(ctx, req)
	p, _ := peer.FromContext(ctx)
	slog.DebugContext(ctx, "gRPC call",
		slog.String("method", info.FullMethod),
		slog.Any("error", err),
		slog.Duration("duration", time.Since(now)),
		slog.Group("peer",
			slog.String("ip", p.Addr.String()),
			slog.Any("auth", p.AuthInfo != nil),
		))
	return resp, err
}

// LogStreamInterceptor logs gRPC streams.
func LogStreamInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	now := time.Now()
	err := handler(srv, ss)
	p, _ := peer.FromContext(ss.Context())
	slog.DebugContext(ss.Context(), "gRPC stream",
		slog.String("method", info.FullMethod),
		slog.Any("error", err),
		slog.Duration("duration", time.Since(now)),
		slog.Group("peer",
			slog.String("ip", p.Addr.String()),
			slog.Any("auth", p.AuthInfo != nil),
		))
	return err
}

// Logger implements grpclog.LoggerV2 to log to slog.
type Logger struct {
	Logger *log.Logger
}

func (l Logger) Info(args ...any)    { l.log("[INFO]", args...) }
func (l Logger) Warning(args ...any) { l.log("[WARN]", args...) }
func (l Logger) Error(args ...any)   { l.log("[ERROR]", args...) }
func (l Logger) Fatal(args ...any)   { l.log("[FATAL]", args...); os.Exit(1) }

func (l Logger) Infoln(args ...any)    { l.logln("[INFO]", args...) }
func (l Logger) Warningln(args ...any) { l.logln("[WARN]", args...) }
func (l Logger) Errorln(args ...any)   { l.logln("[ERROR]", args...) }
func (l Logger) Fatalln(args ...any)   { l.logln("[FATAL]", args...); os.Exit(1) }

func (l Logger) Infof(format string, args ...any)    { l.logf("[INFO] "+format, args...) }
func (l Logger) Warningf(format string, args ...any) { l.logf("[WARN] "+format, args...) }
func (l Logger) Errorf(format string, args ...any)   { l.logf("[ERROR] "+format, args...) }
func (l Logger) Fatalf(format string, args ...any)   { l.logf("[FATAL] "+format, args...); os.Exit(1) }

func (l Logger) log(lvl string, args ...any) {
	l.Logger.Print(append([]any{"[grpclog]", lvl}, args...)...)
}

func (l Logger) logln(lvl string, args ...any) {
	l.Logger.Println(append([]any{"[grpclog]", lvl}, args...)...)
}

func (l Logger) logf(format string, args ...any) {
	l.Logger.Printf("[grpclog] "+format, args...)
}

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
