// Package provider defines the Provider interface and constructor registry for Dynamic DNS providers.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/frizzlabs/frizzlabs-ddns/internal/provider/cloudflare"
	"github.com/frizzlabs/frizzlabs-ddns/internal/provider/duckdns"
	"github.com/frizzlabs/frizzlabs-ddns/internal/provider/noop"
)

// Provider interface defines the contract for updating DNS records for a specific DNS service provider.
type Provider interface {
	// Name returns the identifier of the DNS provider (e.g. "duckdns", "cloudflare").
	Name() string
	// Update updates the DNS provider's record with the target global IPv6 address.
	Update(ctx context.Context, ipv6 netip.Addr) error
}

// New creates a new Provider instance based on the provider name.
func New(name string, domain string, token string, client *http.Client) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "duckdns":
		return duckdns.New(domain, token, client), nil
	case "cloudflare":
		return cloudflare.New(domain, token, client), nil
	case "noop":
		return noop.New(), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", name)
	}
}
