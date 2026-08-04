// Package cloudflare implements the Provider interface for Cloudflare DNS API v4.
package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
)

// Provider implements DNS updates for Cloudflare API v4.
type Provider struct {
	domain  string
	token   string
	client  *http.Client
	baseURL string
}

// New constructs a Cloudflare Provider instance.
func New(domain, token string, client *http.Client) *Provider {
	if client == nil {
		client = http.DefaultClient
	}
	return &Provider{
		domain:  domain,
		token:   token,
		client:  client,
		baseURL: "https://api.cloudflare.com/client/v4",
	}
}

// Name returns the provider name identifier.
func (p *Provider) Name() string {
	return "cloudflare"
}

type cloudflareResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Result []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"result"`
}

// Update searches for the AAAA DNS record on Cloudflare and updates it with the provided IPv6 address.
func (p *Provider) Update(ctx context.Context, ipv6 netip.Addr) error {
	if !ipv6.IsValid() || !ipv6.Is6() {
		return fmt.Errorf("cloudflare: invalid IPv6 address %s", ipv6)
	}

	// 1. Get Zone ID
	zoneID, err := p.getZoneID(ctx)
	if err != nil {
		return fmt.Errorf("cloudflare: %w", err)
	}

	// 2. Get DNS Record ID for AAAA record
	recordID, currentIP, err := p.getDNSRecord(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("cloudflare: %w", err)
	}

	if currentIP == ipv6.String() {
		return nil
	}

	// 3. Update or Create AAAA record
	if recordID != "" {
		return p.updateDNSRecord(ctx, zoneID, recordID, ipv6)
	}
	return p.createDNSRecord(ctx, zoneID, ipv6)
}

func (p *Provider) getZoneID(ctx context.Context) (string, error) {
	reqURL := fmt.Sprintf("%s/zones?name=%s", p.baseURL, url.QueryEscape(p.domain))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating zone request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching zone ID: %w", err)
	}
	defer resp.Body.Close()

	var cfResp cloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return "", fmt.Errorf("decoding zone response: %w", err)
	}

	if !cfResp.Success || len(cfResp.Result) == 0 {
		return "", fmt.Errorf("zone not found for domain %q", p.domain)
	}

	return cfResp.Result[0].ID, nil
}

func (p *Provider) getDNSRecord(ctx context.Context, zoneID string) (recordID string, content string, err error) {
	reqURL := fmt.Sprintf("%s/zones/%s/dns_records?type=AAAA&name=%s", p.baseURL, zoneID, url.QueryEscape(p.domain))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("creating record search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetching DNS record: %w", err)
	}
	defer resp.Body.Close()

	var cfResp cloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return "", "", fmt.Errorf("decoding DNS record response: %w", err)
	}

	if !cfResp.Success {
		return "", "", fmt.Errorf("failed searching DNS records")
	}

	if len(cfResp.Result) > 0 {
		return cfResp.Result[0].ID, cfResp.Result[0].Content, nil
	}

	return "", "", nil
}

func (p *Provider) updateDNSRecord(ctx context.Context, zoneID, recordID string, ipv6 netip.Addr) error {
	reqURL := fmt.Sprintf("%s/zones/%s/dns_records/%s", p.baseURL, zoneID, recordID)
	payload := map[string]interface{}{
		"type":    "AAAA",
		"name":    p.domain,
		"content": ipv6.String(),
		"ttl":     1,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling update payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("creating record update request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending record update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("record update failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (p *Provider) createDNSRecord(ctx context.Context, zoneID string, ipv6 netip.Addr) error {
	reqURL := fmt.Sprintf("%s/zones/%s/dns_records", p.baseURL, zoneID)
	payload := map[string]interface{}{
		"type":    "AAAA",
		"name":    p.domain,
		"content": ipv6.String(),
		"ttl":     1,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling creation payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("creating record request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending record creation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("record creation failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
