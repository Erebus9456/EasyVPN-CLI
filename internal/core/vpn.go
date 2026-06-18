package core

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Erebus9456/EasyVPN-CLI/internal/api"
	"github.com/Erebus9456/EasyVPN-CLI/internal/config"
	"github.com/Erebus9456/EasyVPN-CLI/internal/platform"
	"github.com/Erebus9456/EasyVPN-CLI/internal/state"
	"github.com/Erebus9456/EasyVPN-CLI/internal/ui"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

// Manager coordinates the high-level VPN operations
type Manager struct {
	cfg       *config.Config
	state     *state.Store
	discovery *api.DiscoveryClient
	adapter   platform.PlatformAdapter
}

func NewManager(cfg *config.Config) (*Manager, error) {
	adapter, err := platform.GetAdapter()
	if err != nil {
		return nil, err
	}

	disc, err := api.NewDiscoveryClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{
		cfg:       cfg,
		state:     state.NewStore(cfg.ConfigDir),
		discovery: disc,
		adapter:   adapter,
	}, nil
}

// Connect orchestrates the full provisioning and connection lifecycle
func (m *Manager) Connect(targetServer *models.Server, deviceName string, useKillSwitch bool) error {
	wgCfg, pubKey, err := m.prepareConnection(targetServer, deviceName, true)
	if err != nil {
		return err
	}

	ui.UpdateSpinner("Applying configuration...")
	if err := m.adapter.CreateTunnel(wgCfg); err != nil {
		return err
	}

	newState := &models.VPNState{
		IsConnected:     true,
		ServerID:        targetServer.ID,
		ServerName:      targetServer.Name,
		ServerPublicIP:  targetServer.PublicIP,
		ClientPublicKey: pubKey,
		InterfaceName:   "wg0",
		ConnectedAt:     time.Now(),
		KillSwitch:      useKillSwitch,
	}
	m.state.Save(newState)

	ui.StopSpinnerSuccess(fmt.Sprintf("Successfully connected to %s", targetServer.Name))
	return nil
}

// ExportConfig provisions a peer and writes a WireGuard config file and QR code without activating a tunnel.
func (m *Manager) ExportConfig(targetServer *models.Server, deviceName string, outputPath string) (configPath string, qrPath string, err error) {
	wgCfg, _, err := m.prepareConnection(targetServer, deviceName, false)
	if err != nil {
		return "", "", err
	}

	ui.UpdateSpinner("Writing configuration file...")
	resolvedPath, err := platform.ResolveExportPath(outputPath)
	if err != nil {
		return "", "", err
	}
	if err := platform.WriteWireGuardConfig(resolvedPath, wgCfg); err != nil {
		return "", "", err
	}

	qrPath = platform.QRPathForConfig(resolvedPath)
	ui.UpdateSpinner("Generating QR code...")
	if err := platform.WriteWireGuardQR(qrPath, wgCfg); err != nil {
		return "", "", err
	}

	ui.StopSpinnerSuccess(fmt.Sprintf("Config exported to %s", resolvedPath))
	return resolvedPath, qrPath, nil
}

func (m *Manager) prepareConnection(targetServer *models.Server, deviceName string, overwriteWarning bool) (*models.WireGuardConfig, string, error) {
	currentState, _ := m.state.Load()
	isRotation := false

	if currentState.IsConnected {
		ui.StopSpinnerSuccess("Existing connection detected.")
		msg := "You are already connected to %s. Proceeding will rotate keys on the server. Continue?"
		if overwriteWarning {
			msg = "You are already connected to %s. Proceeding will OVERWRITE your existing config and rotate keys. Continue?"
		}
		if !ui.ConfirmAction(fmt.Sprintf(msg, currentState.ServerName)) {
			return nil, "", models.NewError(models.ErrInternal, "Operation aborted by user", "", nil)
		}
		isRotation = true
	}

	ui.StartSpinner("Verifying system and server health...")
	agent := api.NewAgentClient(targetServer.PublicIP, targetServer.APIPort, m.cfg.ApiToken)
	if err := agent.CheckHealth(); err != nil {
		return nil, "", err
	}

	ui.UpdateSpinner("Generating new secure keypair...")
	privKey, pubKey, err := m.generateKeypair()
	if err != nil {
		return nil, "", err
	}

	var peerResp *models.PeerResponse
	if isRotation && currentState.ClientPublicKey != "" {
		ui.UpdateSpinner("Rotating keys on server...")
		peerResp, err = agent.ReplacePeer(currentState.ClientPublicKey, pubKey)
	} else {
		ui.UpdateSpinner("Provisioning new access...")
		peerResp, err = agent.AddPeer(deviceName, pubKey)
	}
	if err != nil {
		return nil, "", err
	}

	wgCfg := &models.WireGuardConfig{}
	wgCfg.Interface.PrivateKey = privKey
	wgCfg.Interface.Address = peerResp.ClientIP
	wgCfg.Interface.DNS = peerResp.DNS
	wgCfg.Peer.PublicKey = peerResp.ServerPublicKey
	wgCfg.Peer.Endpoint = peerResp.Endpoint
	wgCfg.Peer.AllowedIPs = peerResp.AllowedIPs
	wgCfg.Peer.PersistentKeepalive = 25

	return wgCfg, pubKey, nil
}

// Disconnect cleans up the tunnel and state
func (m *Manager) Disconnect() error {
	ui.StartSpinner("Tearing down tunnel...")

	// Always try to disable kill-switch first to prevent locking out the user
	_ = m.adapter.SetKillSwitch(false, "")

	if err := m.adapter.DestroyTunnel(); err != nil {
		return err
	}

	if err := m.state.Clear(); err != nil {
		return err
	}

	ui.StopSpinnerSuccess("Disconnected successfully")
	return nil
}

// generateKeypair uses the 'wg' binary to create secure keys
func (m *Manager) generateKeypair() (string, string, error) {
	// Generate Private Key
	privCmd := exec.Command("wg", "genkey")
	privOut, err := privCmd.Output()
	if err != nil {
		return "", "", models.NewError(models.ErrInternal, "Failed to generate private key", "Ensure wireguard-tools is installed", err)
	}
	privKey := strings.TrimSpace(string(privOut))

	// Generate Public Key from Private Key
	pubCmd := exec.Command("wg", "pubkey")
	pubCmd.Stdin = strings.NewReader(privKey)
	pubOut, err := pubCmd.Output()
	if err != nil {
		return "", "", models.NewError(models.ErrInternal, "Failed to generate public key", "Ensure wireguard-tools is installed", err)
	}
	pubKey := strings.TrimSpace(string(pubOut))

	return privKey, pubKey, nil
}

// GetAdapter provides access to the platform-specific muscle (for setup and status checks)
func (m *Manager) GetAdapter() (platform.PlatformAdapter, error) {
	if m.adapter == nil {
		return nil, models.NewError(
			models.ErrInternal,
			"Platform adapter not initialized",
			"Please restart the application",
			nil,
		)
	}
	return m.adapter, nil
}
