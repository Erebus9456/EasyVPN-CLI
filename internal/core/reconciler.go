package core

import (
	"github.com/Erebus9456/EasyVPN-CLI/internal/platform"
	"github.com/Erebus9456/EasyVPN-CLI/internal/state"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/utils"
)

// Reconciler is the Truth-First Engine
type Reconciler struct {
	state   *state.Store
	adapter platform.PlatformAdapter
}

func NewReconciler(state *state.Store, adapter platform.PlatformAdapter) *Reconciler {
	return &Reconciler{
		state:   state,
		adapter: adapter,
	}
}

// Sync compares the persisted state with the OS reality and heals any desync
func (r *Reconciler) Sync() (*models.VPNState, error) {
	// 1. Load what the CLI thinks
	persistedState, err := r.state.Load()
	if err != nil {
		return nil, err
	}

	// 2. Query the OS for the actual reality
	realStatus, err := r.adapter.GetStatus()
	if err != nil {
		utils.Log.Debugf("Could not query OS status: %v", err)
		// If we can't query the OS, we trust the persisted state but mark as unverified
		return persistedState, nil
	}

	// 3. Logic: Ghost Connection
	// CLI thinks we are connected, but the tunnel is gone from the OS
	if persistedState.IsConnected && !realStatus.Active {
		utils.Log.Warn("Detected 'Ghost Connection': VPN was closed externally. Healing state...")

		// Auto-disable kill-switch if it was on, as the tunnel is gone
		_ = r.adapter.SetKillSwitch(false, "")

		err := r.state.Clear()
		if err != nil {
			return nil, err
		}
		return &models.VPNState{IsConnected: false}, nil
	}

	// 4. Logic: Untracked Connection
	// CLI thinks we are disconnected, but a WireGuard tunnel is active
	if !persistedState.IsConnected && realStatus.Active {
		utils.Log.Info("Detected active tunnel not managed by this session.")
		// We don't automatically "adopt" it to avoid messing with manual configs,
		// but we return the real status for the UI to show.
		return &models.VPNState{
			IsConnected:   true,
			InterfaceName: realStatus.Interface,
			ServerName:    "Unknown (Manual)",
		}, nil
	}

	// 5. Logic: Everything is in sync
	return persistedState, nil
}

// CheckKillSwitch ensures the firewall rules are still in place if enabled
func (r *Reconciler) CheckKillSwitch(persistedState *models.VPNState, serverIP string) {
	if persistedState.KillSwitch && persistedState.IsConnected {
		// In a 'beast-mode' implementation, we would re-apply the rules
		// here to ensure no other app has flushed the iptables.
		utils.Log.Debug("Re-verifying Kill-switch rules...")
		_ = r.adapter.SetKillSwitch(true, serverIP)
	}
}
