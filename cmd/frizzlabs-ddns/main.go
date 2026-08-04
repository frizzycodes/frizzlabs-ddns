// Command frizzlabs-ddns is a dynamic DNS updater daemon for Linux homelab servers.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/config"
	"github.com/frizzlabs/frizzlabs-ddns/internal/dns"
	"github.com/frizzlabs/frizzlabs-ddns/internal/logger"
	"github.com/frizzlabs/frizzlabs-ddns/internal/network"
	"github.com/frizzlabs/frizzlabs-ddns/internal/provider"
	"github.com/frizzlabs/frizzlabs-ddns/internal/state"
	"github.com/frizzlabs/frizzlabs-ddns/internal/updater"
)

// Build version details set at compile time via ldflags.
var (
	Version   = "1.0.0"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Exit codes following unix conventions for command error taxonomy.
const (
	ExitSuccess       = 0
	ExitConfigError   = 1
	ExitNetworkError  = 2
	ExitStateError    = 3
	ExitProviderError = 4
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("frizzlabs-ddns", flag.ContinueOnError)

	var (
		configPath string
		statePath  string
		dryRun     bool
		verbose    bool
		jsonLog    bool
		showVer    bool
	)

	fs.StringVar(&configPath, "config", "", "Path to JSON configuration file")
	fs.StringVar(&configPath, "c", "", "Path to JSON configuration file (shorthand)")

	fs.StringVar(&statePath, "state", "", "Path to state file (overrides config setting if provided)")
	fs.StringVar(&statePath, "s", "", "Path to state file (shorthand)")

	fs.BoolVar(&dryRun, "dry-run", false, "Simulate execution without modifying DNS or state file")
	fs.BoolVar(&dryRun, "n", false, "Simulate execution without modifying DNS or state file (shorthand)")

	fs.BoolVar(&verbose, "verbose", false, "Enable verbose debug logging")
	fs.BoolVar(&verbose, "v", false, "Enable verbose debug logging (shorthand)")

	fs.BoolVar(&jsonLog, "json", false, "Output log messages in JSON format")

	fs.BoolVar(&showVer, "version", false, "Display version and build metadata")
	fs.BoolVar(&showVer, "V", false, "Display version and build metadata (shorthand)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "frizzlabs-ddns - Dynamic DNS IPv6 Updater Daemon\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  frizzlabs-ddns [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return ExitConfigError
	}

	if showVer {
		fmt.Printf("frizzlabs-ddns v%s (commit: %s, built: %s)\n", Version, Commit, BuildTime)
		return ExitSuccess
	}

	// Initialize structured logger
	log := logger.New(logger.Options{
		Verbose:    verbose,
		JSONFormat: jsonLog,
		Output:     os.Stdout,
	})

	// Resolve configuration file path
	if configPath == "" {
		configPath = findDefaultConfigPath()
	}

	log.Debug("loading configuration", "path", configPath)
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Error("failed loading configuration", "error", err)
		return ExitConfigError
	}

	// Resolve state file path
	if statePath == "" {
		if cfg.StateFile != "" {
			statePath = cfg.StateFile
		} else {
			statePath = findDefaultStatePath()
		}
	}

	// Create root context with signal cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize HTTP client with configured timeout
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
	}

	// Initialize DNS Provider
	prov, err := provider.New(cfg.Provider, cfg.Domain, cfg.Token, httpClient)
	if err != nil {
		log.Error("failed initializing DNS provider", "provider", cfg.Provider, "error", err)
		return ExitConfigError
	}

	// Initialize Network Detector, State Manager, and DNS Resolver
	detector := network.NewDetector()
	stateMgr := state.NewFileManager(statePath)
	sysResolver := dns.NewSystemResolver(nil)

	// Create and run Reconciliation Runner
	runner := updater.NewRunner(cfg, detector, prov, stateMgr, sysResolver, log, dryRun)

	if err := runner.Run(ctx); err != nil {
		log.Error("reconciliation failed", "error", err)
		return mapErrorToExitCode(err)
	}

	return ExitSuccess
}

func findDefaultConfigPath() string {
	candidatePaths := []string{
		"/etc/frizzlabs-ddns/config.json",
		"config.json",
	}
	for _, p := range candidatePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "config.json"
}

func findDefaultStatePath() string {
	systemStatePath := "/var/lib/frizzlabs-ddns/state.json"
	if dir := filepath.Dir(systemStatePath); canWriteToDir(dir) {
		return systemStatePath
	}
	return "state.json"
}

func canWriteToDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	// Try creating a temporary probe file to verify write access
	f, err := os.CreateTemp(dir, ".probe-write-*")
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return true
}

func mapErrorToExitCode(err error) int {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ExitProviderError
	}

	errStr := err.Error()
	switch {
	case containsAny(errStr, "detecting network address", "no global IPv6", "interface"):
		return ExitNetworkError
	case containsAny(errStr, "loading state", "saving state", "state file"):
		return ExitStateError
	case containsAny(errStr, "updating DNS provider", "duckdns", "cloudflare", "http"):
		return ExitProviderError
	default:
		return ExitProviderError
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if len(s) >= len(kw) && (s == kw || filepath.Base(s) == kw || (len(s) > 0 && len(kw) > 0 && containsSubstring(s, kw))) {
			return true
		}
	}
	return false
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsSubstring(s[1:], substr)))
}
