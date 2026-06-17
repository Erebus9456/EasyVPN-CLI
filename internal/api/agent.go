package api

import (
	"fmt"
	"time"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/go-resty/resty/v2"
)

// AgentClient handles communication with a specific VPN Node Agent
type AgentClient struct {
	httpClient *resty.Client
	baseUrl    string
	apiToken   string
}

// NewAgentClient initializes a client for a specific server's IP and API port
func NewAgentClient(ip string, port int, token string) *AgentClient {
	client := resty.New().
		SetTimeout(10*time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(2*time.Second).
		SetHeader("X-API-TOKEN", token).
		SetHeader("Content-Type", "application/json")

	// Base URL format: http://203.0.113.7:5000
	baseUrl := fmt.Sprintf("http://%s:%d", ip, port)

	return &AgentClient{
		httpClient: client,
		baseUrl:    baseUrl,
		apiToken:   token,
	}
}

// CheckHealth verifies if the node agent is online and responding
func (a *AgentClient) CheckHealth() error {
	resp, err := a.httpClient.R().Get(a.baseUrl + "/health")

	if err != nil {
		return models.NewError(
			models.ErrAgentUnreachable,
			"Node agent is not responding",
			"Check if the server is up and the API port is open",
			err,
		)
	}

	if resp.IsError() {
		return models.NewError(
			models.ErrAuthFailed,
			"Node agent health check failed",
			"Ensure your EASYVPN_API_TOKEN is correct",
			fmt.Errorf("status code: %d", resp.StatusCode()),
		)
	}

	return nil
}

// AddPeer requests the server to provision a new WireGuard peer
func (a *AgentClient) AddPeer(deviceName string, publicKey string) (*models.PeerResponse, error) {
	reqBody := models.PeerRequest{
		DeviceName: deviceName,
		PublicKey:  publicKey,
	}

	var peerResp models.PeerResponse
	resp, err := a.httpClient.R().
		SetBody(reqBody).
		SetResult(&peerResp).
		Post(a.baseUrl + "/add-peer")

	if err != nil {
		return nil, models.NewError(
			models.ErrAgentUnreachable,
			"Failed to reach node agent for provisioning",
			"Check server connectivity",
			err,
		)
	}

	if resp.IsError() {
		if resp.StatusCode() == 401 || resp.StatusCode() == 403 {
			return nil, models.NewError(
				models.ErrAuthFailed,
				"Unauthorized access to node agent",
				"Your EASYVPN_API_TOKEN is likely invalid for this server",
				nil,
			)
		}
		return nil, models.NewError(
			models.ErrInternal,
			"Node agent failed to provision peer",
			"The server might be out of internal IP addresses",
			fmt.Errorf("status: %d, body: %s", resp.StatusCode(), resp.String()),
		)
	}

	return &peerResp, nil
}

// ReplacePeer requests the server to swap an old public key for a new one
func (a *AgentClient) ReplacePeer(oldPubKey, newPubKey string) (*models.PeerResponse, error) {
	reqBody := models.ReplacePeerRequest{
		OldPublicKey: oldPubKey,
		PublicKey:    newPubKey,
	}

	var peerResp models.PeerResponse
	resp, err := a.httpClient.R().
		SetBody(reqBody).
		SetResult(&peerResp).
		Post(a.baseUrl + "/replace-peer")

	if err != nil {
		return nil, models.NewError(models.ErrAgentUnreachable, "Failed to reach node agent", "Check server connectivity", err)
	}

	if resp.IsError() {
		return nil, models.NewError(models.ErrInternal, "Node agent failed to rotate keys", resp.String(), nil)
	}

	return &peerResp, nil
}
