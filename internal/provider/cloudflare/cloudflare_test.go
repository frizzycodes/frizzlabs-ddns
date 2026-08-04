package cloudflare_test

import (
	"context"
	"net/netip"
	"testing"

	"github.com/frizzlabs/frizzlabs-ddns/internal/provider/cloudflare"
)

func TestCloudflareProviderName(t *testing.T) {
	p := cloudflare.New("example.com", "token123", nil)
	if p.Name() != "cloudflare" {
		t.Errorf("expected provider name 'cloudflare', got %q", p.Name())
	}
}

func TestCloudflareInvalidAddr(t *testing.T) {
	p := cloudflare.New("example.com", "token123", nil)
	err := p.Update(context.Background(), netip.Addr{})
	if err == nil {
		t.Errorf("expected error for invalid address, got nil")
	}
}
