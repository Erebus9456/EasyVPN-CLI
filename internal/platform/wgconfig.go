package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/skip2/go-qrcode"
)

const defaultExportFilename = "EasyVPN.conf"

// FormatWireGuardConfig renders a standard WireGuard .conf file.
func FormatWireGuardConfig(cfg *models.WireGuardConfig) string {
	return fmt.Sprintf(`[Interface]
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
}

// FormatWireGuardConfigForQR renders the compact format WireGuard mobile apps expect in QR codes.
func FormatWireGuardConfigForQR(cfg *models.WireGuardConfig) string {
	return fmt.Sprintf(`[Interface]
PrivateKey=%s
Address=%s
DNS=%s

[Peer]
PublicKey=%s
Endpoint=%s
AllowedIPs=%s
PersistentKeepalive=%d
`, cfg.Interface.PrivateKey, cfg.Interface.Address, cfg.Interface.DNS,
		cfg.Peer.PublicKey, cfg.Peer.Endpoint, cfg.Peer.AllowedIPs, cfg.Peer.PersistentKeepalive)
}

// QRPathForConfig derives a PNG path from the exported .conf path.
func QRPathForConfig(configPath string) string {
	ext := filepath.Ext(configPath)
	if ext == "" {
		return configPath + ".png"
	}
	return strings.TrimSuffix(configPath, ext) + ".png"
}

// DefaultExportPath returns the default export location in the current working directory.
func DefaultExportPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return defaultExportFilename, nil
	}
	return filepath.Join(cwd, defaultExportFilename), nil
}

// ResolveExportPath expands ~ and treats directory paths as export folders.
func ResolveExportPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return DefaultExportPath()
	}

	path := trimmed
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", models.NewError(models.ErrInternal, "Failed to resolve home directory", "", err)
		}
		path = strings.Replace(path, "~", home, 1)
	}

	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		path = filepath.Join(path, defaultExportFilename)
	}

	return path, nil
}

// WriteWireGuardConfig writes a WireGuard config file to the given path.
func WriteWireGuardConfig(path string, cfg *models.WireGuardConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return models.NewError(models.ErrInternal, "Failed to create export directory", "Check folder permissions", err)
	}
	content := FormatWireGuardConfig(cfg)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return models.NewError(models.ErrInternal, "Failed to write WireGuard config", "Check folder permissions", err)
	}
	return nil
}

// WriteWireGuardQR writes a scannable QR code PNG for WireGuard mobile clients.
func WriteWireGuardQR(path string, cfg *models.WireGuardConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return models.NewError(models.ErrInternal, "Failed to create QR export directory", "Check folder permissions", err)
	}
	payload := FormatWireGuardConfigForQR(cfg)
	if err := qrcode.WriteFile(payload, qrcode.Medium, 512, path); err != nil {
		return models.NewError(models.ErrInternal, "Failed to write WireGuard QR code", "Check folder permissions", err)
	}
	return nil
}
