package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/frizzlabs/frizzlabs-ddns/internal/config"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name: "valid config with interface",
			cfg: config.Config{
				Version:   1,
				Provider:  "duckdns",
				Interface: "eth0",
				Domain:    "test",
				Token:     "secret-token",
			},
			wantErr: false,
		},
		{
			name: "valid config with matchDefaultRoute",
			cfg: config.Config{
				Version:           1,
				Provider:          "cloudflare",
				MatchDefaultRoute: true,
				Domain:            "test",
				Token:             "secret-token",
			},
			wantErr: false,
		},
		{
			name: "missing version",
			cfg: config.Config{
				Version:   0,
				Interface: "eth0",
				Domain:    "test",
				Token:     "secret-token",
			},
			wantErr: true,
		},
		{
			name: "missing domain",
			cfg: config.Config{
				Version:   1,
				Interface: "eth0",
				Domain:    "",
				Token:     "secret-token",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			cfg: config.Config{
				Version:   1,
				Interface: "eth0",
				Domain:    "test",
				Token:     "",
			},
			wantErr: true,
		},
		{
			name: "unsupported provider",
			cfg: config.Config{
				Version:   1,
				Provider:  "unknown-provider",
				Interface: "eth0",
				Domain:    "test",
				Token:     "secret-token",
			},
			wantErr: true,
		},
		{
			name: "no interface and matchDefaultRoute false",
			cfg: config.Config{
				Version:           1,
				Provider:          "duckdns",
				Interface:         "",
				MatchDefaultRoute: false,
				Domain:            "test",
				Token:             "secret-token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	validConfigPath := filepath.Join(tempDir, "config.json")

	content := `{
		"version": 1,
		"provider": "duckdns",
		"interface": "enp4s0",
		"domain": "homelab",
		"token": "token123",
		"timeoutSec": 15
	}`

	if err := os.WriteFile(validConfigPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed writing test config file: %v", err)
	}

	cfg, err := config.Load(validConfigPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Domain != "homelab" {
		t.Errorf("expected Domain 'homelab', got %q", cfg.Domain)
	}
	if cfg.TimeoutSec != 15 {
		t.Errorf("expected TimeoutSec 15, got %d", cfg.TimeoutSec)
	}
}
