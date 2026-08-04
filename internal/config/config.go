// Package config provides configuration parsing, validation, and schema definitions.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DefaultTimeoutSec specifies the default HTTP timeout in seconds.
const DefaultTimeoutSec = 10

// Config defines the structure of the JSON configuration file for frizzlabs-ddns.
type Config struct {
	// Version specifies the configuration file format version (must be >= 1).
	Version int `json:"version"`

	// Provider specifies the DNS provider name (e.g., "duckdns", "cloudflare", "noop").
	Provider string `json:"provider"`

	// Interface specifies the target network interface name (e.g., "enp4s0").
	Interface string `json:"interface,omitempty"`

	// MatchDefaultRoute automatically selects the interface owning the default IPv6 route when true.
	MatchDefaultRoute bool `json:"matchDefaultRoute"`

	// Domain specifies the sub-domain or host name to update.
	Domain string `json:"domain"`

	// Token specifies the authentication token or API key for the provider.
	Token string `json:"token"`

	// StateFile specifies an optional custom override path for state.json.
	StateFile string `json:"stateFile,omitempty"`

	// VerifyDNS performs DNS AAAA resolution to detect and repair DNS drift when true (defaults to true).
	VerifyDNS *bool `json:"verifyDNS,omitempty"`

	// TimeoutSec specifies the HTTP client timeout in seconds (defaults to 10s).
	TimeoutSec int `json:"timeoutSec,omitempty"`
}

// IsVerifyDNSEnabled returns true if VerifyDNS is enabled (defaults to true if unassigned).
func (c *Config) IsVerifyDNSEnabled() bool {
	if c.VerifyDNS == nil {
		return true
	}
	return *c.VerifyDNS
}

// Load reads and parses a JSON configuration file from the specified path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing JSON config from %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config in %q: %w", path, err)
	}

	return &cfg, nil
}

// Validate checks the configuration for required fields, sane defaults, and valid values.
func (c *Config) Validate() error {
	if c.Version < 1 {
		return fmt.Errorf("config version must be at least 1, got %d", c.Version)
	}

	if strings.TrimSpace(c.Domain) == "" {
		return fmt.Errorf("field 'domain' is required and cannot be empty")
	}

	if strings.TrimSpace(c.Token) == "" {
		return fmt.Errorf("field 'token' is required and cannot be empty")
	}

	c.Provider = strings.ToLower(strings.TrimSpace(c.Provider))
	if c.Provider == "" {
		c.Provider = "duckdns"
	}

	switch c.Provider {
	case "duckdns", "cloudflare", "noop":
		// Supported providers
	default:
		return fmt.Errorf("unsupported DNS provider %q (must be 'duckdns', 'cloudflare', or 'noop')", c.Provider)
	}

	if strings.TrimSpace(c.Interface) == "" && !c.MatchDefaultRoute {
		return fmt.Errorf("must specify either 'interface' name or set 'matchDefaultRoute': true")
	}

	if c.TimeoutSec <= 0 {
		c.TimeoutSec = DefaultTimeoutSec
	}

	return nil
}
