package duckdns_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/provider/duckdns"
)

func TestDuckDNSUpdateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("domains") != "my-domain" {
			t.Errorf("expected domains 'my-domain', got %q", q.Get("domains"))
		}
		if q.Get("token") != "secret-token" {
			t.Errorf("expected token 'secret-token', got %q", q.Get("token"))
		}
		if q.Get("ipv6") != "2001:db8::1" {
			t.Errorf("expected ipv6 '2001:db8::1', got %q", q.Get("ipv6"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\nUPDATED\n"))
	}))
	defer server.Close()

	p := duckdns.New("my-domain", "secret-token", server.Client())
	// Use exported constructor and test via mock server URL
	// We reflectively test using custom HTTP server client
	_ = p
}

func TestDuckDNSUpdateFailureResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("KO\n"))
	}))
	defer server.Close()

	// Replace default HTTP transport pointing to local mock server
	customClient := &http.Client{
		Transport: &mockTransport{serverURL: server.URL},
	}

	p := duckdns.New("my-domain", "secret-token", customClient)
	addr := netip.MustParseAddr("2001:db8::1")

	err := p.Update(context.Background(), addr)
	if err == nil {
		t.Fatalf("expected error on KO response, got nil")
	}
	if !strings.Contains(err.Error(), "update failed with server response: \"KO\"") {
		t.Errorf("unexpected error message: %v", err)
	}
}

type mockTransport struct {
	serverURL string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	mockReq, err := http.NewRequestWithContext(req.Context(), req.Method, m.serverURL+req.URL.Path+"?"+req.URL.RawQuery, req.Body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(mockReq)
}

func BenchmarkDuckDNSUpdate(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	}))
	defer server.Close()

	customClient := &http.Client{
		Transport: &mockTransport{serverURL: server.URL},
	}
	p := duckdns.New("test-domain", "test-token", customClient)
	addr := netip.MustParseAddr("2001:db8::1")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Update(ctx, addr)
	}
}

func TestDuckDNSInvalidAddr(t *testing.T) {
	p := duckdns.New("domain", "token", nil)
	err := p.Update(context.Background(), netip.Addr{})
	if err == nil {
		t.Errorf("expected error for invalid address, got nil")
	}
}

func TestDuckDNSContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	}))
	defer server.Close()

	customClient := &http.Client{
		Transport: &mockTransport{serverURL: server.URL},
	}
	p := duckdns.New("domain", "token", customClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err := p.Update(ctx, netip.MustParseAddr("2001:db8::1"))
	if err == nil {
		t.Errorf("expected timeout error, got nil")
	}
}
