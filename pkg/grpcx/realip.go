package grpcx

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
	"fmt"
)

// RealIP extracts the real IP address from the context.
func RealIP(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	var firstIP string
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := md.Get(h)
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			realIP := net.ParseIP(ip)
			if firstIP == "" && realIP != nil {
				firstIP = ip
			}
			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				continue
			}
			return ip, nil
		}
	}

	if firstIP != "" {
		return firstIP, nil
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		ip, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			return "", fmt.Errorf("can't parse ip %q: %w", p.Addr.String(), err)
		}
		if netIP := net.ParseIP(ip); netIP != nil {
			return ip, nil
		}
	}

	return "", fmt.Errorf("no valid ip found")
}

func mustParseCIDR(s string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Errorf("must parse cidr %q: %w", s, err))
	}
	return ipnet
}

var privateSubnets = []*net.IPNet{
	mustParseCIDR("10.0.0.0/8"),
	mustParseCIDR("100.64.0.0/10"),
	mustParseCIDR("172.16.0.0/12"),
	mustParseCIDR("192.0.0.0/24"),
	mustParseCIDR("192.168.0.0/16"),
	mustParseCIDR("198.18.0.0/15"),
	mustParseCIDR("::1/128"),
	mustParseCIDR("fc00::/7"),
	mustParseCIDR("fe80::/10"),
}

func isPrivateSubnet(ip net.IP) bool {
	for _, ipnet := range privateSubnets {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}
