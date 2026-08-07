// Package dns provides DNS resolution functionality using Go's standard net.Resolver.
package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// ErrNotFound is returned when a DNS hostname or AAAA record does not exist.
var ErrNotFound = errors.New("DNS AAAA record not found")

// Resolver defines the interface for querying DNS records.
type Resolver interface {
	// LookupAAAA resolves AAAA (IPv6) records for the specified hostname.
	LookupAAAA(ctx context.Context, host string) ([]netip.Addr, error)
}

// IsNotFound returns true if the error indicates that the DNS hostname or AAAA record does not exist.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound
	}
	return false
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
		if IsNotFound(err) {
			return nil, fmt.Errorf("lookup AAAA for %q: %w", trimmedHost, ErrNotFound)
		}
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
		return nil, fmt.Errorf("lookup AAAA for %q: %w", trimmedHost, ErrNotFound)
	}

	return ipv6Addrs, nil
}
