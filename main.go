// Package main is an application entrypoint.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/Semior001/grpc-echo/echopb"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc"
	"net"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/grpclog"
	"github.com/Semior001/grpc-echo/pkg/grpcx"
	"github.com/Semior001/grpc-echo/pkg/service"
)

var opts struct {
	Addr  string `short:"a" long:"addr" env:"ADDR" default:":8080" description:"Address to listen on"`
	JSON  bool   `long:"json"           env:"JSON"             description:"Enable JSON logging"`
	Debug bool   `long:"debug"          env:"DEBUG"            description:"Enable debug mode"`
}

var version = "unknown"

func getVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok && version == "unknown" {
		return bi.Main.Version
	}
	return version
}

func main() {
	_, _ = fmt.Fprintf(os.Stderr, "grpc-echo %s\n", getVersion())

	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	setupLog(opts.Debug, opts.JSON)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		sig := <-stop
		slog.Warn("caught signal", slog.Any("signal", sig))
		cancel()
	}()

	if err := run(ctx); err != nil {
		slog.Error("failed to start service", slog.Any("error", err))
	}
}

func run(ctx context.Context) error {
	svc := &service.EchoService{}
	healthHandler := health.NewServer()
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			svc.AppendTimestampInterceptor,
			grpcx.LogInterceptor,
		),
	)

	grpclog.SetLoggerV2(grpcx.Logger{})

	healthpb.RegisterHealthServer(srv, healthHandler)
	echopb.RegisterEchoServiceServer(srv, svc)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", opts.Addr, err)
	}

	ewg, ctx := errgroup.WithContext(ctx)
	ewg.Go(func() error {
		healthHandler.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
		if err := srv.Serve(lis); err != nil {
			return fmt.Errorf("proxy server: %w", err)
		}
		return nil
	})
	ewg.Go(func() error {
		<-ctx.Done()
		healthHandler.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		srv.GracefulStop()
		return nil
	})

	if err := ewg.Wait(); err != nil {
		return err
	}

	return nil
}

func setupLog(dbg, json bool) {
	defer slog.Info("prepared logger", slog.Bool("debug", dbg), slog.Bool("json", json))

	handlerOpts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if dbg {
		handlerOpts.Level = slog.LevelDebug
		handlerOpts.AddSource = true
	}

	var handler slog.Handler
	if json {
		handler = slog.NewJSONHandler(os.Stderr, handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, handlerOpts)
	}

	slog.SetDefault(slog.New(handler))
}
