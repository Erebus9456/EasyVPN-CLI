package models

import "time"

// Server represents a VPN gateway retrieved from Supabase discovery
// pkg/models/server.go

type Server struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Region             string    `json:"region"`
	PublicIP           string    `json:"public_ip"`
	EndpointPort       int       `json:"endpoint_port"`
	WireguardPublicKey string    `json:"wireguard_public_key"` // Matches your SQL
	Status             string    `json:"status"`               // "online", "offline", etc
	LastHeartbeat      time.Time `json:"last_heartbeat"`
	CreatedAt          time.Time `json:"created_at"`

	// Default the API Port to 5000 if not in DB yet
	APIPort int `json:"api_port"`
}

// PeerRequest is the payload sent to the Node Agent to provision a new client
type PeerRequest struct {
	DeviceName string `json:"name"`
	PublicKey  string `json:"public_key"`
}

// PeerResponse is the data returned by the Node Agent after successful provisioning
type PeerResponse struct {
	ClientIP        string `json:"client_ip"`
	ServerPublicKey string `json:"server_public_key"`
	Endpoint        string `json:"endpoint"` // e.g., "203.0.113.7:51820"
	DNS             string `json:"dns"`
	AllowedIPs      string `json:"allowed_ips"`
}

// WireGuardConfig holds all the data needed to generate a local .conf file
type WireGuardConfig struct {
	Interface struct {
		PrivateKey string
		Address    string
		DNS        string
	}
	Peer struct {
		PublicKey           string
		Endpoint            string
		AllowedIPs          string
		PersistentKeepalive int
	}
}

// ReplacePeerRequest is the payload for key rotation
type ReplacePeerRequest struct {
	OldPublicKey string `json:"old_public_key"`
	PublicKey    string `json:"public_key"`
}

// VPNState is the schema for our state.json file to track the current connection
// Update VPNState to include the ClientPublicKey
type VPNState struct {
	IsConnected     bool      `json:"is_connected"`
	ServerID        string    `json:"server_id"`
	ServerName      string    `json:"server_name"`
	ServerPublicIP  string    `json:"server_public_ip"`  // Added to help Kill-switch cleanup
	ClientPublicKey string    `json:"client_public_key"` // Added for rotation
	InterfaceName   string    `json:"interface_name"`
	ConnectedAt     time.Time `json:"connected_at"`
	KillSwitch      bool      `json:"kill_switch_enabled"`
}
