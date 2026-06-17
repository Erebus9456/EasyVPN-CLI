package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

type WindowsAdapter struct{}

func NewWindowsAdapter() *WindowsAdapter {
	return &WindowsAdapter{}
}

func (a *WindowsAdapter) Name() string { return "windows" }

// IsPrivileged checks for Administrative rights on Windows
func (a *WindowsAdapter) IsPrivileged() (bool, error) {
	// On Windows, the most reliable way to check for Admin is to attempt
	// to open a protected system resource.
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (a *WindowsAdapter) CheckRequirements() []models.EasyVPNError {
	var errs []models.EasyVPNError
	// On Windows, we look for the official WireGuard path or the command in PATH
	_, err := exec.LookPath("wireguard.exe")
	if err != nil {
		errs = append(errs, *models.NewError(
			models.ErrMissingDependency,
			"WireGuard for Windows not found",
			"Please install the official WireGuard client from wireguard.com",
			err,
		))
	}
	return errs
}

func (a *WindowsAdapter) InstallDependencies() error {
	return models.NewError(
		models.ErrInternal,
		"Automatic installation not supported on Windows",
		"Please download and install WireGuard from https://www.wireguard.com/install/",
		nil,
	)
}

func (a *WindowsAdapter) CreateTunnel(cfg *models.WireGuardConfig) error {
	// 1. Generate the config file
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

	// 2. Write to config directory (WireGuard on Windows usually expects configs in specific places or passed via CLI)
	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".easyvpn", "wg0.conf")
	if err := os.WriteFile(configPath, []byte(confContent), 0600); err != nil {
		return models.NewError(models.ErrInternal, "Failed to write Windows WG config", "Check folder permissions", err)
	}

	// 3. Command WireGuard to install and start the tunnel as a service
	// wireguard.exe /installmanager "path\to\config.conf"
	cmd := exec.Command("wireguard.exe", "/installmanager", configPath)
	if err := cmd.Run(); err != nil {
		return models.NewError(models.ErrTunnelFailed, "Failed to start WireGuard service on Windows", "Ensure you are running as Administrator", err)
	}

	return nil
}

func (a *WindowsAdapter) DestroyTunnel() error {
	// On Windows, we tell the installmanager to uninstall the specific tunnel service
	// The service name is usually based on the filename (wg0)
	cmd := exec.Command("wireguard.exe", "/uninstallmanager", "wg0")
	if err := cmd.Run(); err != nil {
		// If the tunnel doesn't exist, ignore the error
		return nil
	}
	return nil
}

func (a *WindowsAdapter) SetKillSwitch(enabled bool, serverIP string) error {
	// On Windows, we use 'netsh' to manage the Advanced Firewall
	if enabled {
		// Block all outbound except to the VPN server and the local loopback
		exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=EasyVPN-Lock", "dir=out", "action=block").Run()
		exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=EasyVPN-Allow-Server", "dir=out", "action=allow", "remoteip="+serverIP).Run()
	} else {
		// Clean up rules
		exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=EasyVPN-Lock").Run()
		exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=EasyVPN-Allow-Server").Run()
	}
	return nil
}

func (a *WindowsAdapter) GetStatus() (*TunnelStatus, error) {
	// Query the list of running services to see if WireGuard Tunnel: wg0 is active
	cmd := exec.Command("sc", "query", "WireGuardTunnel$wg0")
	output, err := cmd.Output()
	if err != nil {
		return &TunnelStatus{Active: false}, nil
	}

	if strings.Contains(string(output), "RUNNING") {
		return &TunnelStatus{Active: true, Interface: "wg0"}, nil
	}

	return &TunnelStatus{Active: false}, nil
}
