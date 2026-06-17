package platform

import (
	"runtime"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

// TunnelStatus represents the real-time state of the OS network interface
type TunnelStatus struct {
	Active    bool
	Interface string
	LocalIP   string
	PublicKey string
	BytesSent int64
	BytesRecv int64
}

// PlatformAdapter defines the set of actions every OS must implement
type PlatformAdapter interface {
	// Name returns the OS name (linux, windows, darwin)
	Name() string

	// IsPrivileged checks if the app is running as root/admin
	IsPrivileged() (bool, error)

	// CheckRequirements verifies if wireguard-tools and necessary utilities are installed
	CheckRequirements() []models.EasyVPNError

	// InstallDependencies attempts to install missing tools (Option 1 in menu)
	InstallDependencies() error

	// CreateTunnel sets up the WireGuard interface and routing
	CreateTunnel(cfg *models.WireGuardConfig) error

	// DestroyTunnel tears down the interface and restores routing
	DestroyTunnel() error

	// SetKillSwitch enables or disables the firewall-based leak protection
	SetKillSwitch(enabled bool, serverIP string) error

	// GetStatus queries the OS directly to see if the tunnel is actually running
	GetStatus() (*TunnelStatus, error)
}

// GetAdapter returns the correct implementation based on the user's OS
func GetAdapter() (PlatformAdapter, error) {
	switch runtime.GOOS {
	case "linux":
		return NewLinuxAdapter(), nil
	case "windows":
		return NewWindowsAdapter(), nil
	case "darwin":
		return NewDarwinAdapter(), nil
	default:
		return nil, models.NewError(
			models.ErrInternal,
			"Unsupported Operating System",
			"EasyVPN currently supports Linux, Windows, and macOS",
			nil,
		)
	}
}
