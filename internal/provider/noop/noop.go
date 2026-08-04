// Package noop provides a no-op implementation of Provider for testing, benchmark, and dry-run execution.
package noop

import (
	"context"
	"fmt"
	"net/netip"
)

// Provider implements Provider with no external side-effects.
type Provider struct{}

// New constructs a NoOp Provider instance.
func New() *Provider {
	return &Provider{}
}

// Name returns the provider name identifier.
func (p *Provider) Name() string {
	return "noop"
}

// Update simulates a successful DNS record update.
func (p *Provider) Update(ctx context.Context, ipv6 netip.Addr) error {
	if !ipv6.IsValid() {
		return fmt.Errorf("noop: invalid IPv6 address")
	}
	return nil
}
