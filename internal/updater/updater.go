// Package updater orchestrates network address detection, DNS resolution reconciliation, and provider record synchronization with retries.
package updater

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/config"
	"github.com/frizzlabs/frizzlabs-ddns/internal/dns"
	"github.com/frizzlabs/frizzlabs-ddns/internal/network"
	"github.com/frizzlabs/frizzlabs-ddns/internal/provider"
	"github.com/frizzlabs/frizzlabs-ddns/internal/state"
)

// Runner coordinates address resolution, DNS drift verification, and DNS provider updates.
type Runner struct {
	cfg         *config.Config
	detector    network.Detector
	provider    provider.Provider
	stateMgr    state.Manager
	dnsResolver dns.Resolver
	logger      *slog.Logger
	dryRun      bool
}

// NewRunner constructs a Runner with dependency injection.
func NewRunner(cfg *config.Config, det network.Detector, prov provider.Provider, stMgr state.Manager, resolver dns.Resolver, l *slog.Logger, dryRun bool) *Runner {
	return &Runner{
		cfg:         cfg,
		detector:    det,
		provider:    prov,
		stateMgr:    stMgr,
		dnsResolver: resolver,
		logger:      l,
		dryRun:      dryRun,
	}
}

// Run executes a single reconciliation pass comparing local IPv6, cached state, and live DNS AAAA records.
func (r *Runner) Run(ctx context.Context) error {
	r.logger.Info("starting Dynamic DNS IPv6 reconciliation engine", "provider", r.provider.Name(), "domain", r.cfg.Domain, "dryRun", r.dryRun)

	// 1. Detect current global IPv6 address
	currentIP, err := r.detector.DetectIPv6(ctx, r.cfg.Interface, r.cfg.MatchDefaultRoute)
	if err != nil {
		return fmt.Errorf("detecting network address: %w", err)
	}

	r.logger.Info("detected current global IPv6 address", "ipv6", currentIP.String())

	// 2. Load cached state
	st, err := r.stateMgr.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	cachedIP, _ := st.GetLastIPv6()

	// 3. Resolve live DNS AAAA record if verifyDNS is enabled
	var (
		dnsResolved       bool
		dnsMatchesCurrent bool
		dnsIsMissing      bool
		resolvedIP        netip.Addr
	)

	if r.cfg.IsVerifyDNSEnabled() && r.dnsResolver != nil {
		hostToResolve := r.resolveHostname()
		r.logger.Debug("querying live AAAA record for drift verification", "host", hostToResolve)

		addrs, err := r.dnsResolver.LookupAAAA(ctx, hostToResolve)
		if err != nil {
			if dns.IsNotFound(err) {
				// Authoritative response: record is missing / cleared / NXDOMAIN on DNS server!
				dnsResolved = true
				dnsMatchesCurrent = false
				dnsIsMissing = true
				r.logger.Warn("DNS record missing/cleared (NXDOMAIN) on DNS server", "host", hostToResolve)
			} else {
				// Transient network error (e.g. timeout, connection refused) -> fallback to cached state
				r.logger.Warn("unable to resolve AAAA record due to network error, falling back to cached state", "host", hostToResolve, "error", err)
			}
		} else if len(addrs) > 0 {
			dnsResolved = true
			resolvedIP = addrs[0]
			for _, a := range addrs {
				if a == currentIP {
					dnsMatchesCurrent = true
					break
				}
			}
			r.logger.Debug("resolved live AAAA record", "host", hostToResolve, "resolvedIPv6", resolvedIP.String(), "matchesCurrent", dnsMatchesCurrent)
		}
	}

	// 4. Decision Matrix: Compare Current, Cached, and DNS state
	ipChanged := !cachedIP.IsValid() || cachedIP != currentIP
	dnsDrift := dnsResolved && !dnsMatchesCurrent

	switch {
	case !ipChanged && dnsResolved && dnsMatchesCurrent:
		r.logger.Info("DNS record already synchronized", "ipv6", currentIP.String(), "lastUpdated", st.LastUpdated.Format(time.RFC3339))
		return nil

	case !ipChanged && !dnsResolved:
		r.logger.Info("IPv6 address has not changed, update skipped (DNS verification unconfirmed)", "ipv6", currentIP.String())
		return nil

	case !ipChanged && dnsDrift:
		if dnsIsMissing {
			r.logger.Warn("DNS drift detected! Hostname record is missing/cleared (NXDOMAIN) on DNS server", "currentIPv6", currentIP.String(), "cachedIPv6", cachedIP.String())
		} else {
			r.logger.Warn("DNS drift detected!", "resolvedAAAA", resolvedIP.String(), "currentIPv6", currentIP.String(), "cachedIPv6", cachedIP.String())
		}
		r.logger.Info("repairing DNS record...")

	case ipChanged:
		r.logger.Info("machine IPv6 address change detected", "oldIPv6", cachedIP.String(), "newIPv6", currentIP.String())
	}

	if r.dryRun {
		r.logger.Info("[DRY-RUN] Would update DNS provider and save state", "provider", r.provider.Name(), "newIPv6", currentIP.String())
		return nil
	}

	// 5. Execute Provider Update with Retries (250ms, 500ms, 1000ms)
	if err := r.updateWithRetry(ctx, currentIP); err != nil {
		return fmt.Errorf("updating DNS provider %s: %w", r.provider.Name(), err)
	}

	// 6. Update and persist state
	st.Provider = r.provider.Name()
	st.SetLastIPv6(currentIP)
	st.LastUpdated = time.Now().UTC()

	if err := r.stateMgr.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	r.logger.Info("successfully updated DNS record and synchronized state", "ipv6", currentIP.String())
	return nil
}

func (r *Runner) resolveHostname() string {
	domain := strings.TrimSpace(r.cfg.Domain)
	if r.provider.Name() == "duckdns" && !strings.Contains(domain, ".") {
		return domain + ".duckdns.org"
	}
	return domain
}

// updateWithRetry attempts provider Update with backoff delays of 250ms, 500ms, 1000ms.
func (r *Runner) updateWithRetry(ctx context.Context, ipv6 netip.Addr) error {
	backoffs := []time.Duration{
		250 * time.Millisecond,
		500 * time.Millisecond,
		1000 * time.Millisecond,
	}

	var lastErr error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := r.provider.Update(ctx, ipv6)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < len(backoffs) {
			delay := backoffs[attempt]
			r.logger.Warn("provider update failed, retrying...", "attempt", attempt+1, "delay", delay.String(), "error", err.Error())

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("all update attempts failed; last error: %w", lastErr)
}
