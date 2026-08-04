package noop_test

import (
	"context"
	"net/netip"
	"testing"

	"github.com/frizzlabs/frizzlabs-ddns/internal/provider/noop"
)

func TestNoOpProvider(t *testing.T) {
	p := noop.New()
	if p.Name() != "noop" {
		t.Errorf("expected 'noop', got %q", p.Name())
	}

	addr := netip.MustParseAddr("2001:db8::1")
	if err := p.Update(context.Background(), addr); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
