// Package state provides atomic persistence and loading of IP resolution state using JSON.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"time"
)

// CurrentStateVersion specifies the state file format version.
const CurrentStateVersion = 1

// State represents cached dynamic DNS IP resolution status.
type State struct {
	// Version specifies the state file schema version.
	Version int `json:"version"`

	// Provider specifies the name of the DNS provider used for the last successful update.
	Provider string `json:"provider"`

	// LastIPv6 contains the string representation of the last successfully set global IPv6 address.
	LastIPv6 string `json:"lastIPv6,omitempty"`

	// LastUpdated records the UTC timestamp of the last successful update.
	LastUpdated time.Time `json:"lastUpdated"`
}

// Manager defines the interface for state persistence operations.
type Manager interface {
	Load() (*State, error)
	Save(st *State) error
}

// FileManager manages state loading and atomic writing to a file on disk.
type FileManager struct {
	filePath string
}

// NewFileManager constructs a new FileManager for the given target state file path.
func NewFileManager(filePath string) *FileManager {
	return &FileManager{filePath: filePath}
}

// GetLastIPv6 parses and returns the cached LastIPv6 string as a netip.Addr.
func (s *State) GetLastIPv6() (netip.Addr, error) {
	if s.LastIPv6 == "" {
		return netip.Addr{}, nil
	}
	addr, err := netip.ParseAddr(s.LastIPv6)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("parsing cached IPv6 %q: %w", s.LastIPv6, err)
	}
	return addr, nil
}

// SetLastIPv6 updates the state's LastIPv6 string from a valid netip.Addr.
func (s *State) SetLastIPv6(addr netip.Addr) {
	if addr.IsValid() {
		s.LastIPv6 = addr.String()
	} else {
		s.LastIPv6 = ""
	}
}

// Load reads and parses the JSON state file. If the file does not exist, an empty initial State is returned without error.
func (m *FileManager) Load() (*State, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{
				Version: CurrentStateVersion,
			}, nil
		}
		return nil, fmt.Errorf("reading state file %q: %w", m.filePath, err)
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parsing JSON state from %q: %w", m.filePath, err)
	}

	return &st, nil
}

// Save atomically writes state data to disk by writing to a temporary file, performing fsync, and renaming.
func (m *FileManager) Save(st *State) error {
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state directory %q: %w", dir, err)
	}

	st.Version = CurrentStateVersion

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state JSON: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "state-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary state file in %q: %w", dir, err)
	}
	tmpName := tmpFile.Name()

	// Ensure temp file is cleaned up if rename fails
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if err := tmpFile.Chmod(0600); err != nil {
		return fmt.Errorf("setting permissions on temporary state file: %w", err)
	}

	if _, err := io.WriteString(tmpFile, string(data)+"\n"); err != nil {
		return fmt.Errorf("writing state data: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing state file to disk: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary state file: %w", err)
	}

	if err := os.Rename(tmpName, m.filePath); err != nil {
		return fmt.Errorf("renaming temporary state file to %q: %w", m.filePath, err)
	}

	return nil
}
