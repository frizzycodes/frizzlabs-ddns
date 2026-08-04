// Package duckdns implements the Provider interface for DuckDNS (https://www.duckdns.org).
package duckdns

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

// Provider implements DNS updates for DuckDNS.
type Provider struct {
	domain  string
	token   string
	client  *http.Client
	baseURL string
}

// New constructs a DuckDNS Provider instance.
func New(domain, token string, client *http.Client) *Provider {
	if client == nil {
		client = http.DefaultClient
	}
	return &Provider{
		domain:  domain,
		token:   token,
		client:  client,
		baseURL: "https://www.duckdns.org/update",
	}
}

// Name returns the provider name identifier.
func (p *Provider) Name() string {
	return "duckdns"
}

// Update sends an HTTP GET request to DuckDNS to update the IPv6 record.
func (p *Provider) Update(ctx context.Context, ipv6 netip.Addr) error {
	if !ipv6.IsValid() || !ipv6.Is6() {
		return fmt.Errorf("duckdns: invalid IPv6 address %s", ipv6)
	}

	params := url.Values{}
	params.Set("domains", p.domain)
	params.Set("token", p.token)
	params.Set("ipv6", ipv6.String())
	params.Set("verbose", "true")

	reqURL := fmt.Sprintf("%s?%s", p.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("duckdns: creating HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "frizzlabs-ddns/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("duckdns: sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("duckdns: unexpected HTTP status code %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("duckdns: reading response body: %w", err)
		}
		return fmt.Errorf("duckdns: empty response received from server")
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "OK" {
		return fmt.Errorf("duckdns: update failed with server response: %q", firstLine)
	}

	return nil
}
