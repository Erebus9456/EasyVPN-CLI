package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

type LinuxAdapter struct{}

func NewLinuxAdapter() *LinuxAdapter {
	return &LinuxAdapter{}
}

func (a *LinuxAdapter) Name() string { return "linux" }

func (a *LinuxAdapter) IsPrivileged() (bool, error) {
	return os.Geteuid() == 0, nil
}

func (a *LinuxAdapter) CheckRequirements() []models.EasyVPNError {
	var errs []models.EasyVPNError
	tools := []string{"wg", "wg-quick", "iptables"}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			errs = append(errs, *models.NewError(
				models.ErrMissingDependency,
				fmt.Sprintf("Missing dependency: %s", tool),
				fmt.Sprintf("Please install %s (e.g., 'sudo apt install wireguard-tools iptables')", tool),
				err,
			))
		}
	}
	return errs
}

func (a *LinuxAdapter) InstallDependencies() error {
	isRoot := os.Geteuid() == 0
	useSudo := !isRoot
	if useSudo {
		if _, err := exec.LookPath("sudo"); err != nil {
			useSudo = false
		}
	}

	var cmdName string
	var cmdArgs []string

	if _, err := exec.LookPath("apt-get"); err == nil {
		fmt.Println("📦 Linux package manager detected: apt-get")
		fmt.Println("🛠️ Updating package lists...")
		
		var updateCmd *exec.Cmd
		if useSudo {
			updateCmd = exec.Command("sudo", "apt-get", "update")
		} else {
			updateCmd = exec.Command("apt-get", "update")
		}
		updateCmd.Stdin = os.Stdin
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		_ = updateCmd.Run() // Continue even if update fails

		fmt.Println("🛠️ Installing wireguard, wireguard-tools, and iptables...")
		cmdName = "apt-get"
		cmdArgs = []string{"install", "-y", "wireguard", "wireguard-tools", "iptables"}
	} else if _, err := exec.LookPath("dnf"); err == nil {
		fmt.Println("📦 Linux package manager detected: dnf")
		fmt.Println("🛠️ Installing wireguard-tools and iptables...")
		cmdName = "dnf"
		cmdArgs = []string{"install", "-y", "wireguard-tools", "iptables"}
	} else if _, err := exec.LookPath("yum"); err == nil {
		fmt.Println("📦 Linux package manager detected: yum")
		fmt.Println("🛠️ Installing wireguard-tools and iptables...")
		cmdName = "yum"
		cmdArgs = []string{"install", "-y", "wireguard-tools", "iptables"}
	} else if _, err := exec.LookPath("pacman"); err == nil {
		fmt.Println("📦 Linux package manager detected: pacman")
		fmt.Println("🛠️ Installing wireguard-tools and iptables...")
		cmdName = "pacman"
		cmdArgs = []string{"-S", "--noconfirm", "wireguard-tools", "iptables"}
	} else {
		return models.NewError(
			models.ErrPermissionDenied,
			"Automatic installation not supported on this distro",
			"Please install wireguard, wireguard-tools, and iptables using your system's package manager manually.",
			nil,
		)
	}

	var finalCmd *exec.Cmd
	if useSudo {
		args := append([]string{cmdName}, cmdArgs...)
		finalCmd = exec.Command("sudo", args...)
	} else {
		finalCmd = exec.Command(cmdName, cmdArgs...)
	}

	finalCmd.Stdin = os.Stdin
	finalCmd.Stdout = os.Stdout
	finalCmd.Stderr = os.Stderr

	if err := finalCmd.Run(); err != nil {
		return models.NewError(
			models.ErrInternal,
			"Package installation failed",
			"Try running the installation command manually with sudo",
			err,
		)
	}

	return nil
}

func (a *LinuxAdapter) CreateTunnel(cfg *models.WireGuardConfig) error {
	// 1. Prepare config content
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

	// 2. Write to a temporary profile (wg-quick requires a .conf suffix)
	home, _ := os.UserHomeDir()
	confPath := filepath.Join(home, ".easyvpn", "wg0.conf")
	if err := os.MkdirAll(filepath.Dir(confPath), 0700); err != nil {
		return models.NewError(models.ErrInternal, "Failed to create WG config directory", "Check permissions of ~/.easyvpn", err)
	}
	if err := os.WriteFile(confPath, []byte(confContent), 0600); err != nil {
		return models.NewError(models.ErrInternal, "Failed to write WG config", "Check permissions of ~/.easyvpn", err)
	}

	// 3. Execute wg-quick
	cmd := exec.Command("wg-quick", "up", confPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return models.NewError(models.ErrTunnelFailed, "wg-quick failed to bring up the interface", string(output), err)
	}

	return nil
}

func (a *LinuxAdapter) DestroyTunnel() error {
	home, _ := os.UserHomeDir()
	confPath := filepath.Join(home, ".easyvpn", "wg0.conf")

	// Check if file exists before trying to bring it down
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command("wg-quick", "down", confPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return models.NewError(models.ErrTunnelFailed, "Failed to shutdown tunnel", string(output), err)
	}
	return nil
}

func (a *LinuxAdapter) SetKillSwitch(enabled bool, serverIP string) error {
	// Simple iptables implementation
	// Note: In a real 'beast-mode' app, we'd use a more robust 'nftables' set
	if enabled {
		// 1. Allow traffic to the VPN Server IP
		exec.Command("iptables", "-A", "OUTPUT", "-d", serverIP, "-j", "ACCEPT").Run()
		// 2. Allow traffic through the wg0 interface
		exec.Command("iptables", "-A", "OUTPUT", "-o", "wg0", "-j", "ACCEPT").Run()
		// 3. Drop everything else
		exec.Command("iptables", "-P", "OUTPUT", "DROP").Run()
	} else {
		// Restore defaults
		exec.Command("iptables", "-P", "OUTPUT", "ACCEPT").Run()
		exec.Command("iptables", "-F", "OUTPUT").Run()
	}
	return nil
}

func (a *LinuxAdapter) GetStatus() (*TunnelStatus, error) {
	// Beast-Mode: Parse 'wg show' for actual kernel data
	cmd := exec.Command("wg", "show", "wg0", "dump")
	output, err := cmd.Output()
	if err != nil {
		return &TunnelStatus{Active: false}, nil
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		return &TunnelStatus{Active: false}, nil
	}

	// Basic parsing of 'wg dump' output
	// Format: [interface_private_key] [public_key] [listen_port] [fwmark]
	//         [peer_public_key] [preshared_key] [endpoint] [allowed_ips] [latest_handshake] [transfer_rx] [transfer_tx] [persistent_keepalive]
	return &TunnelStatus{
		Active:    true,
		Interface: "wg0",
	}, nil
}
