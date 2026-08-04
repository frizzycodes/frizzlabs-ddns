package state_test

import (
	"net/netip"
	"path/filepath"
	"testing"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/state"
)

func TestStateLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "non_existent_state.json")

	mgr := state.NewFileManager(statePath)
	st, err := mgr.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}

	if st.Version != state.CurrentStateVersion {
		t.Errorf("expected version %d, got %d", state.CurrentStateVersion, st.Version)
	}
	if st.LastIPv6 != "" {
		t.Errorf("expected empty LastIPv6, got %q", st.LastIPv6)
	}
}

func TestStateSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.json")

	mgr := state.NewFileManager(statePath)
	ip := netip.MustParseAddr("2001:db8::ffff")
	now := time.Now().UTC().Truncate(time.Second)

	initialState := &state.State{
		Provider:    "duckdns",
		LastUpdated: now,
	}
	initialState.SetLastIPv6(ip)

	if err := mgr.Save(initialState); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	loadedState, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loadedState.Provider != "duckdns" {
		t.Errorf("expected Provider 'duckdns', got %q", loadedState.Provider)
	}

	loadedIP, err := loadedState.GetLastIPv6()
	if err != nil {
		t.Fatalf("GetLastIPv6() failed: %v", err)
	}

	if loadedIP != ip {
		t.Errorf("expected IP %s, got %s", ip, loadedIP)
	}
}

func BenchmarkStateSave(b *testing.B) {
	tempDir := b.TempDir()
	statePath := filepath.Join(tempDir, "bench_state.json")
	mgr := state.NewFileManager(statePath)

	st := &state.State{
		Provider:    "duckdns",
		LastIPv6:    "2001:db8::1",
		LastUpdated: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.Save(st)
	}
}
