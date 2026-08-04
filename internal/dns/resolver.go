// Package dns provides DNS resolution functionality using Go's standard net.Resolver.
package dns

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// Resolver defines the interface for querying DNS records.
type Resolver interface {
	// LookupAAAA resolves AAAA (IPv6) records for the specified hostname.
	LookupAAAA(ctx context.Context, host string) ([]netip.Addr, error)
}

// SystemResolver implements Resolver using Go's standard library net.Resolver.
type SystemResolver struct {
	resolver *net.Resolver
}

// NewSystemResolver constructs a SystemResolver instance.
func NewSystemResolver(customResolver *net.Resolver) *SystemResolver {
	if customResolver == nil {
		customResolver = net.DefaultResolver
	}
	return &SystemResolver{resolver: customResolver}
}

// LookupAAAA queries AAAA records for the given hostname and returns valid IPv6 addresses.
func (s *SystemResolver) LookupAAAA(ctx context.Context, host string) ([]netip.Addr, error) {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		return nil, fmt.Errorf("lookup AAAA: empty host specified")
	}

	netAddrs, err := s.resolver.LookupNetIP(ctx, "ip6", trimmedHost)
	if err != nil {
		return nil, fmt.Errorf("lookup AAAA for %q: %w", trimmedHost, err)
	}

	var ipv6Addrs []netip.Addr
	for _, addr := range netAddrs {
		addr = addr.Unmap()
		if addr.IsValid() && addr.Is6() {
			ipv6Addrs = append(ipv6Addrs, addr)
		}
	}

	if len(ipv6Addrs) == 0 {
		return nil, fmt.Errorf("no valid AAAA records found for host %q", trimmedHost)
	}

	return ipv6Addrs, nil
}
