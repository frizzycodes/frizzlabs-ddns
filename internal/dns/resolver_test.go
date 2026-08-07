package dns_test

import (
	"context"
	"errors"
	"net"
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
	if dns.IsNotFound(err) {
		t.Errorf("context cancellation error should not be classified as IsNotFound")
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
	if !dns.IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) == true for non-existent domain, got false (err: %v)", err)
	}
	if !strings.Contains(err.Error(), "lookup AAAA") {
		t.Errorf("unexpected error format: %v", err)
	}
}

func TestIsNotFoundHelper(t *testing.T) {
	if dns.IsNotFound(nil) {
		t.Errorf("IsNotFound(nil) should be false")
	}
	if !dns.IsNotFound(dns.ErrNotFound) {
		t.Errorf("IsNotFound(ErrNotFound) should be true")
	}

	netErrNotFound := &net.DNSError{IsNotFound: true}
	if !dns.IsNotFound(netErrNotFound) {
		t.Errorf("IsNotFound(&net.DNSError{IsNotFound: true}) should be true")
	}

	netErrTimeout := &net.DNSError{IsTimeout: true, IsNotFound: false}
	if dns.IsNotFound(netErrTimeout) {
		t.Errorf("IsNotFound(&net.DNSError{IsTimeout: true}) should be false")
	}

	genericErr := errors.New("connection reset by peer")
	if dns.IsNotFound(genericErr) {
		t.Errorf("IsNotFound(genericErr) should be false")
	}
}
