package state

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

// Store handles the persistence of VPN state to disk
type Store struct {
	FilePath string
}

// NewStore creates a new state manager pointing to the config directory
func NewStore(configDir string) *Store {
	return &Store{
		FilePath: filepath.Join(configDir, "state.json"),
	}
}

// Save persists the current VPN state to disk atomically
func (s *Store) Save(state *models.VPNState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return models.NewError(models.ErrInternal, "Failed to serialize state", "Contact support", err)
	}

	// Beast-Mode: Atomic Write
	// 1. Write to a temporary file
	tmpFile := s.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return models.NewError(models.ErrInternal, "Failed to write temp state file", "Check disk space and permissions", err)
	}

	// 2. Rename temp file to actual file (Atomic on most OSs)
	if err := os.Rename(tmpFile, s.FilePath); err != nil {
		return models.NewError(models.ErrInternal, "Failed to finalize state file", "Check file locks", err)
	}

	return nil
}

// Load retrieves the VPN state from disk
func (s *Store) Load() (*models.VPNState, error) {
	if _, err := os.Stat(s.FilePath); os.IsNotExist(err) {
		// No state file exists, return an empty "disconnected" state
		return &models.VPNState{IsConnected: false}, nil
	}

	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		return nil, models.NewError(models.ErrInternal, "Failed to read state file", "Check file permissions", err)
	}

	var state models.VPNState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, models.NewError(models.ErrInternal, "Corrupt state file", "Run 'easyvpn disconnect' to reset", err)
	}

	return &state, nil
}

// Clear resets the state to disconnected and updates the file
func (s *Store) Clear() error {
	return s.Save(&models.VPNState{
		IsConnected: false,
	})
}
