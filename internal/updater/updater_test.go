package updater_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/netip"
	"testing"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/config"
	"github.com/frizzlabs/frizzlabs-ddns/internal/state"
	"github.com/frizzlabs/frizzlabs-ddns/internal/updater"
)

type mockDetector struct {
	addr netip.Addr
	err  error
}

func (m *mockDetector) DetectIPv6(ctx context.Context, interfaceName string, matchDefaultRoute bool) (netip.Addr, error) {
	return m.addr, m.err
}

type mockProvider struct {
	name      string
	updateErr error
	calls     int
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Update(ctx context.Context, ipv6 netip.Addr) error {
	m.calls++
	return m.updateErr
}

type mockStateMgr struct {
	st  *state.State
	err error
}

func (m *mockStateMgr) Load() (*state.State, error) { return m.st, m.err }
func (m *mockStateMgr) Save(st *state.State) error  { m.st = st; return nil }

type mockResolver struct {
	addrs []netip.Addr
	err   error
	calls int
}

func (m *mockResolver) LookupAAAA(ctx context.Context, host string) ([]netip.Addr, error) {
	m.calls++
	return m.addrs, m.err
}

func boolPtr(b bool) *bool { return &b }

func TestRunnerNoChangeAndDNSMatches(t *testing.T) {
	addr := netip.MustParseAddr("2001:db8::1")
	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(true)}
	det := &mockDetector{addr: addr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(addr)
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{addrs: []netip.Addr{addr}}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, false)

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if prov.calls != 0 {
		t.Errorf("expected 0 provider calls when IP matches DNS and state, got %d", prov.calls)
	}
	if res.calls != 1 {
		t.Errorf("expected 1 DNS lookup call, got %d", res.calls)
	}
}

func TestRunnerMachineIPChanged(t *testing.T) {
	oldAddr := netip.MustParseAddr("2001:db8::1")
	newAddr := netip.MustParseAddr("2001:db8::2")

	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(true)}
	det := &mockDetector{addr: newAddr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(oldAddr)
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{addrs: []netip.Addr{oldAddr}}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, false)

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if prov.calls != 1 {
		t.Errorf("expected 1 provider call, got %d", prov.calls)
	}

	lastIP, _ := stMgr.st.GetLastIPv6()
	if lastIP != newAddr {
		t.Errorf("expected updated state IP %s, got %s", newAddr, lastIP)
	}
}

func TestRunnerDNSDriftDetectedAndRepaired(t *testing.T) {
	currentAddr := netip.MustParseAddr("2001:db8::1")
	driftedDNSAddr := netip.MustParseAddr("2001:db8::9999") // External tool or manual edit modified DNS

	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(true)}
	det := &mockDetector{addr: currentAddr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(currentAddr) // Local state matches current machine IP
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{addrs: []netip.Addr{driftedDNSAddr}} // Live DNS differs!

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, false)

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected no error during DNS drift repair, got %v", err)
	}

	if prov.calls != 1 {
		t.Errorf("expected 1 provider update call to repair DNS drift, got %d", prov.calls)
	}
}

func TestRunnerDNSLookupUnavailableFallback(t *testing.T) {
	addr := netip.MustParseAddr("2001:db8::1")

	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(true)}
	det := &mockDetector{addr: addr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(addr)
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{err: errors.New("DNS query timeout")}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, false)

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected execution success on DNS lookup fallback, got %v", err)
	}

	if prov.calls != 0 {
		t.Errorf("expected 0 provider calls when DNS lookup fails and state matches, got %d", prov.calls)
	}
}

func TestRunnerVerifyDNSDisabled(t *testing.T) {
	addr := netip.MustParseAddr("2001:db8::1")

	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(false)} // Disabled
	det := &mockDetector{addr: addr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(addr)
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{addrs: []netip.Addr{netip.MustParseAddr("2001:db8::9999")}}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, false)

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected execution success, got %v", err)
	}

	if res.calls != 0 {
		t.Errorf("expected 0 DNS lookup calls when VerifyDNS is disabled, got %d", res.calls)
	}
}

func TestRunnerDryRunDriftRepair(t *testing.T) {
	currentAddr := netip.MustParseAddr("2001:db8::1")
	driftedDNSAddr := netip.MustParseAddr("2001:db8::9999")

	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(true)}
	det := &mockDetector{addr: currentAddr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(currentAddr)
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{addrs: []netip.Addr{driftedDNSAddr}}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, true) // DryRun = true

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected no error in dry-run drift repair, got %v", err)
	}

	if prov.calls != 0 {
		t.Errorf("expected 0 provider calls in dry-run, got %d", prov.calls)
	}
}

func TestRunnerProviderRetryFailure(t *testing.T) {
	newAddr := netip.MustParseAddr("2001:db8::2")

	cfg := &config.Config{Domain: "test.com", VerifyDNS: boolPtr(true)}
	det := &mockDetector{addr: newAddr}
	prov := &mockProvider{name: "duckdns", updateErr: errors.New("network error")}
	st := &state.State{}
	stMgr := &mockStateMgr{st: st}
	res := &mockResolver{addrs: []netip.Addr{netip.MustParseAddr("2001:db8::1")}}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, res, logger, false)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := runner.Run(ctx)
	if err == nil {
		t.Fatalf("expected error after retries failure, got nil")
	}

	if prov.calls != 4 {
		t.Errorf("expected 4 provider attempts, got %d", prov.calls)
	}
}
