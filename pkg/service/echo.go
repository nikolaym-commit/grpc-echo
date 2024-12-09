// Package service contains echo service implementation.
package service

import (
	"github.com/Semior001/grpc-echo/echopb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"google.golang.org/grpc"
	"time"
	"context"
	"google.golang.org/grpc/peer"
)

// EchoService implements the EchoServiceServer interface.
type EchoService struct{ echopb.UnimplementedEchoServiceServer }

// Echo returns the request as a response with some additional timestamps.
func (s *EchoService) Echo(ctx context.Context, req *echopb.EchoRequest) (resp *echopb.EchoResponse, err error) {
	md, _ := metadata.FromIncomingContext(ctx)
	resp = &echopb.EchoResponse{
		Headers:          make(map[string]string, len(md)),
		Body:             req.Ping,
		HandlerReachedAt: timestamppb.Now(),
	}
	if p, ok := peer.FromContext(ctx); ok {
		resp.RemoteAddr = p.Addr.String()
	}
	for k, vals := range md {
		resp.Headers[k] = strings.Join(vals, ",")
	}
	defer func() { resp.HandlerRespondedAt = timestamppb.Now() }()
	return resp, nil
}

// AppendTimestampInterceptor appends timestamps to the echo response.
func (*EchoService) AppendTimestampInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	recvAt := time.Now()
	if resp, err = handler(ctx, req); err != nil {
		return resp, err
	}

	eresp, ok := resp.(*echopb.EchoResponse)
	if !ok {
		return resp, err
	}

	eresp.ReceivedAt = timestamppb.New(recvAt)
	eresp.SentAt = timestamppb.Now()

	return eresp, nil
}
