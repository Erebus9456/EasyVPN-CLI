package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

type DarwinAdapter struct{}

func NewDarwinAdapter() *DarwinAdapter {
	return &DarwinAdapter{}
}

func (a *DarwinAdapter) Name() string { return "darwin" }

func (a *DarwinAdapter) IsPrivileged() (bool, error) {
	// For export-only mode, we don't strictly need root,
	// so we return true to pass the check.
	return true, nil
}

func (a *DarwinAdapter) CheckRequirements() []models.EasyVPNError {
	var errs []models.EasyVPNError
	// On macOS, we only need 'wg' to generate keys for the export
	if _, err := exec.LookPath("wg"); err != nil {
		errs = append(errs, *models.NewError(
			models.ErrMissingDependency,
			"WireGuard Tools missing",
			"Required for key generation. Install via Option 1.",
			err,
		))
	}
	return errs
}

func (a *DarwinAdapter) InstallDependencies() error {
	// 1. Check if Homebrew is installed
	if _, err := exec.LookPath("brew"); err != nil {
		return models.NewError(
			models.ErrMissingDependency,
			"Homebrew not found",
			"Please install Homebrew first from https://brew.sh",
			err,
		)
	}

	// 2. Run brew install
	fmt.Println("🍺 Installing wireguard-tools via Homebrew...")
	cmd := exec.Command("brew", "install", "wireguard-tools")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return models.NewError(
			models.ErrInternal,
			"Homebrew installation failed",
			"Try running 'brew install wireguard-tools' manually in terminal",
			err,
		)
	}

	return nil
}

func (a *DarwinAdapter) CreateTunnel(cfg *models.WireGuardConfig) error {
	home, _ := os.UserHomeDir()
	exportPath := filepath.Join(home, "Desktop", "EasyVPN_macOS.conf")
	return WriteWireGuardConfig(exportPath, cfg)
}

func (a *DarwinAdapter) DestroyTunnel() error {
	// Not applicable for macOS CLI
	return nil
}

func (a *DarwinAdapter) SetKillSwitch(enabled bool, serverIP string) error {
	return models.NewError(
		models.ErrInternal,
		"Kill-switch via CLI is not supported on macOS",
		"Please use the 'Kill-switch' feature inside the official WireGuard macOS app settings.",
		nil,
	)
}

func (a *DarwinAdapter) GetStatus() (*TunnelStatus, error) {
	// Since we don't manage the tunnel, we always report as inactive
	return &TunnelStatus{Active: false}, nil
}
