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
	// 1. Generate the WireGuard config content
	confContent := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = %d
`, cfg.Interface.PrivateKey, cfg.Interface.Address, cfg.Interface.DNS,
		cfg.Peer.PublicKey, cfg.Peer.Endpoint, cfg.Peer.AllowedIPs, cfg.Peer.PersistentKeepalive)

	// 2. Define the export path (Desktop is convenient for macOS users)
	home, _ := os.UserHomeDir()
	exportPath := filepath.Join(home, "Desktop", "EasyVPN_macOS.conf")

	// 3. Write the file
	if err := os.WriteFile(exportPath, []byte(confContent), 0600); err != nil {
		return models.NewError(models.ErrInternal, "Failed to export macOS config", "Check Desktop write permissions", err)
	}

	// 4. Print instructions to the user (Beast-Mode UX)
	fmt.Println("\n🍏 macOS Export Successful!")
	fmt.Printf("📂 Config saved to: %s\n", exportPath)
	fmt.Println("--------------------------------------------------")
	fmt.Println("How to use:")
	fmt.Println("1. Open the official 'WireGuard' app on your Mac.")
	fmt.Println("2. Click 'Import tunnel(s) from file' (or use the + icon).")
	fmt.Println("3. Select the 'EasyVPN_macOS.conf' file from your Desktop.")
	fmt.Println("4. Click 'Activate'.")
	fmt.Println("--------------------------------------------------")

	return nil
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
