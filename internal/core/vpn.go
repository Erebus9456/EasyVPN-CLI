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
	// 1. Check existing state for "Already Connected"
	currentState, _ := m.state.Load()
	isRotation := false

	if currentState.IsConnected {
		ui.StopSpinnerSuccess("Existing connection detected.")
		msg := fmt.Sprintf("You are already connected to %s. Proceeding will OVERWRITE your existing config and rotate keys. Continue?", currentState.ServerName)
		if !ui.ConfirmAction(msg) {
			return models.NewError(models.ErrInternal, "Connection aborted by user", "", nil)
		}
		isRotation = true
	}

	// 2. Pre-flight and Health Checks
	ui.StartSpinner("Verifying system and server health...")
	agent := api.NewAgentClient(targetServer.PublicIP, targetServer.APIPort, m.cfg.ApiToken)
	if err := agent.CheckHealth(); err != nil {
		return err
	}

	// 3. Generate NEW WireGuard Keys
	ui.UpdateSpinner("Generating new secure keypair...")
	privKey, pubKey, err := m.generateKeypair()
	if err != nil {
		return err
	}

	// 4. Provision or Replace Peer
	var peerResp *models.PeerResponse
	if isRotation && currentState.ClientPublicKey != "" {
		ui.UpdateSpinner("Rotating keys on server...")
		peerResp, err = agent.ReplacePeer(currentState.ClientPublicKey, pubKey)
	} else {
		ui.UpdateSpinner("Provisioning new access...")
		peerResp, err = agent.AddPeer(deviceName, pubKey)
	}

	if err != nil {
		return err
	}

	// 5. Build and Apply Config (Logic remains the same)
	wgCfg := &models.WireGuardConfig{}
	wgCfg.Interface.PrivateKey = privKey
	wgCfg.Interface.Address = peerResp.ClientIP
	wgCfg.Interface.DNS = peerResp.DNS
	wgCfg.Peer.PublicKey = peerResp.ServerPublicKey
	wgCfg.Peer.Endpoint = peerResp.Endpoint
	wgCfg.Peer.AllowedIPs = peerResp.AllowedIPs
	wgCfg.Peer.PersistentKeepalive = 25

	ui.UpdateSpinner("Applying configuration...")
	if err := m.adapter.CreateTunnel(wgCfg); err != nil {
		return err
	}

	// 6. Finalize State
	newState := &models.VPNState{
		IsConnected:     true,
		ServerID:        targetServer.ID,
		ServerName:      targetServer.Name,
		ServerPublicIP:  targetServer.PublicIP,
		ClientPublicKey: pubKey, // Store the new public key for future rotations
		InterfaceName:   "wg0",
		ConnectedAt:     time.Now(),
		KillSwitch:      useKillSwitch,
	}
	m.state.Save(newState)

	ui.StopSpinnerSuccess(fmt.Sprintf("Successfully connected to %s", targetServer.Name))
	return nil
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
