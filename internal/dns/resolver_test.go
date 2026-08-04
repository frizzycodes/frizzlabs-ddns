package dns_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/dns"
)

func TestSystemResolverEmptyHost(t *testing.T) {
	r := dns.NewSystemResolver(nil)
	_, err := r.LookupAAAA(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty host, got nil")
	}
}

func TestSystemResolverContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := dns.NewSystemResolver(nil)
	_, err := r.LookupAAAA(ctx, "google.com")
	if err == nil {
		t.Errorf("expected error for cancelled context, got nil")
	}
}

func TestSystemResolverNonExistentDomain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r := dns.NewSystemResolver(nil)
	_, err := r.LookupAAAA(ctx, "non-existent-subdomain-123456789.example.invalid")
	if err == nil {
		t.Errorf("expected resolution failure error for invalid domain, got nil")
	}
	if !strings.Contains(err.Error(), "lookup AAAA") {
		t.Errorf("unexpected error format: %v", err)
	}
}
