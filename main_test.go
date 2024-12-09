package main

import (
	"testing"
	"math/rand"
	"os"
	"syscall"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"fmt"
	"time"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"context"
	"github.com/Semior001/grpc-echo/echopb"
	"reflect"
)

func TestMain_run(t *testing.T) {
	port, conn := setup(t)
	defer conn.Close()
	waitForServerUp(t, conn)

	now := time.Now()

	client := echopb.NewEchoServiceClient(conn)
	resp, err := client.Echo(context.Background(), &echopb.EchoRequest{Ping: "hello"})
	if err != nil {
		t.Fatalf("failed to check health: %v", err)
	}

	t.Logf("response: %+v", resp)
	assert(t, resp.Body == "hello", "unexpected response body: %+v", resp.Body)
	assert(t, reflect.DeepEqual(map[string]string{
		":authority":   fmt.Sprintf("localhost:%d", port),
		"content-type": "application/grpc",
		"user-agent":   fmt.Sprintf("grpc-echo-test-ua grpc-go/%s", grpc.Version),
	}, resp.Headers), "unexpected headers: %+v", resp.Headers)

	const threshold = time.Millisecond * 10
	assert(t, time.Since(now) < threshold, "more than %s passed: %s", threshold, time.Since(now))
	assert(t, resp.ReceivedAt.AsTime().Sub(now) < threshold,
		"received_at: all ts must be within %s: %s", threshold, resp.ReceivedAt.AsTime().Sub(now))
	assert(t, resp.SentAt.AsTime().Sub(now) < threshold,
		"sent_at: all ts must be within %s: %s", threshold, resp.SentAt.AsTime().Sub(now))
	assert(t, resp.HandlerReachedAt.AsTime().Sub(now) < threshold,
		"handler_reached_at: all ts must be within %s: %s", threshold, resp.HandlerReachedAt.AsTime().Sub(now))
	assert(t, resp.HandlerRespondedAt.AsTime().Sub(now) < threshold,
		"handler_responded_at: all ts must be within %s: %s", threshold,
		resp.HandlerRespondedAt.AsTime().Sub(now))
}

func assert(tb testing.TB, cond bool, format string, args ...any) {
	tb.Helper()
	if !cond {
		tb.Fatalf(format, args...)
	}
}

func setup(tb testing.TB) (port int, conn *grpc.ClientConn) {
	port = 40000 + int(rand.Int31n(10000))
	os.Args = []string{"test", "--addr", ":" + strconv.Itoa(port)}

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if e != nil {
			tb.Errorf("failed to send signal: %v", e)
		}
	}()

	started, finished := make(chan struct{}), make(chan struct{})
	go func() {
		tb.Logf("running server on port %d", port)
		close(started)
		main()
		close(finished)
	}()

	tb.Cleanup(func() {
		close(done)
		<-finished
	})

	<-started
	time.Sleep(time.Millisecond * 50) // do not start right away
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUserAgent("grpc-echo-test-ua"))
	if err != nil {
		tb.Fatalf("failed to create client: %v", err)
	}

	return port, conn
}

func waitForServerUp(tb testing.TB, conn *grpc.ClientConn) {
	healthClient := healthpb.NewHealthClient(conn)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		st, err := healthClient.Check(context.Background(), &healthpb.HealthCheckRequest{})
		if err == nil && st.Status == healthpb.HealthCheckResponse_SERVING {
			return
		}
	}
	tb.Fatal("server is not up")
}
