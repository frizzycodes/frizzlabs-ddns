// Package updater orchestrates network address detection, state reconciliation, and provider record synchronization with retries.
package updater

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/config"
	"github.com/frizzlabs/frizzlabs-ddns/internal/network"
	"github.com/frizzlabs/frizzlabs-ddns/internal/provider"
	"github.com/frizzlabs/frizzlabs-ddns/internal/state"
)

// Runner coordinates address resolution, comparison with stored state, and DNS provider updates.
type Runner struct {
	cfg      *config.Config
	detector network.Detector
	provider provider.Provider
	stateMgr state.Manager
	logger   *slog.Logger
	dryRun   bool
}

// NewRunner constructs a Runner with dependency injection.
func NewRunner(cfg *config.Config, det network.Detector, prov provider.Provider, stMgr state.Manager, l *slog.Logger, dryRun bool) *Runner {
	return &Runner{
		cfg:      cfg,
		detector: det,
		provider: prov,
		stateMgr: stMgr,
		logger:   l,
		dryRun:   dryRun,
	}
}

// Run executes a single reconciliation pass.
func (r *Runner) Run(ctx context.Context) error {
	r.logger.Info("starting Dynamic DNS IPv6 reconciliation", "provider", r.provider.Name(), "domain", r.cfg.Domain, "dryRun", r.dryRun)

	// 1. Detect current global IPv6 address
	detectedIP, err := r.detector.DetectIPv6(ctx, r.cfg.Interface, r.cfg.MatchDefaultRoute)
	if err != nil {
		return fmt.Errorf("detecting network address: %w", err)
	}

	r.logger.Info("detected current global IPv6 address", "ipv6", detectedIP.String())

	// 2. Load cached state
	st, err := r.stateMgr.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	cachedIP, _ := st.GetLastIPv6()

	// 3. Compare detected IP with cached IP
	if cachedIP.IsValid() && cachedIP == detectedIP {
		r.logger.Info("IPv6 address has not changed, update skipped", "ipv6", detectedIP.String(), "lastUpdated", st.LastUpdated.Format(time.RFC3339))
		return nil
	}

	r.logger.Info("IPv6 address change detected", "oldIPv6", cachedIP.String(), "newIPv6", detectedIP.String())

	if r.dryRun {
		r.logger.Info("[DRY-RUN] Would update DNS provider and save state", "provider", r.provider.Name(), "newIPv6", detectedIP.String())
		return nil
	}

	// 4. Update DNS Provider with 250ms, 500ms, 1000ms retries
	if err := r.updateWithRetry(ctx, detectedIP); err != nil {
		return fmt.Errorf("updating DNS provider %s: %w", r.provider.Name(), err)
	}

	// 5. Save updated state
	st.Provider = r.provider.Name()
	st.SetLastIPv6(detectedIP)
	st.LastUpdated = time.Now().UTC()

	if err := r.stateMgr.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	r.logger.Info("successfully updated DNS record and saved state", "ipv6", detectedIP.String())
	return nil
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
