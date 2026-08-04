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

func TestRunnerNoChange(t *testing.T) {
	addr := netip.MustParseAddr("2001:db8::1")
	cfg := &config.Config{Domain: "test.com"}
	det := &mockDetector{addr: addr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(addr)
	stMgr := &mockStateMgr{st: st}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, logger, false)

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if prov.calls != 0 {
		t.Errorf("expected 0 provider calls when IP hasn't changed, got %d", prov.calls)
	}
}

func TestRunnerIPChanged(t *testing.T) {
	oldAddr := netip.MustParseAddr("2001:db8::1")
	newAddr := netip.MustParseAddr("2001:db8::2")

	cfg := &config.Config{Domain: "test.com"}
	det := &mockDetector{addr: newAddr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	st.SetLastIPv6(oldAddr)
	stMgr := &mockStateMgr{st: st}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, logger, false)

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

func TestRunnerDryRun(t *testing.T) {
	newAddr := netip.MustParseAddr("2001:db8::2")

	cfg := &config.Config{Domain: "test.com"}
	det := &mockDetector{addr: newAddr}
	prov := &mockProvider{name: "duckdns"}
	st := &state.State{}
	stMgr := &mockStateMgr{st: st}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, logger, true) // dryRun = true

	err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("expected no error in dry-run, got %v", err)
	}

	if prov.calls != 0 {
		t.Errorf("expected 0 provider calls in dry-run, got %d", prov.calls)
	}
}

func TestRunnerProviderRetryFailure(t *testing.T) {
	newAddr := netip.MustParseAddr("2001:db8::2")

	cfg := &config.Config{Domain: "test.com"}
	det := &mockDetector{addr: newAddr}
	prov := &mockProvider{name: "duckdns", updateErr: errors.New("network error")}
	st := &state.State{}
	stMgr := &mockStateMgr{st: st}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := updater.NewRunner(cfg, det, prov, stMgr, logger, false)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := runner.Run(ctx)
	if err == nil {
		t.Fatalf("expected error after retries failure, got nil")
	}

	if prov.calls != 4 { // Initial attempt + 3 retries = 4 total attempts
		t.Errorf("expected 4 provider attempts, got %d", prov.calls)
	}
}
