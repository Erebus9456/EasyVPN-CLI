package api

import (
	"github.com/Erebus9456/EasyVPN-CLI/internal/config"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/go-resty/resty/v2"
)

type IPClient struct {
	httpClient *resty.Client
	url        string
}

func NewIPClient(cfg *config.Config) *IPClient {
	return &IPClient{
		httpClient: resty.New(),
		url:        cfg.PublicIpCheckUrl,
	}
}

func (c *IPClient) GetPublicIP() (string, error) {
	var result struct {
		IP string `json:"ip"`
	}

	resp, err := c.httpClient.R().
		SetResult(&result).
		Get(c.url)

	if err != nil || resp.IsError() {
		return "", models.NewError(models.ErrAgentUnreachable, "Failed to fetch public IP", "Check your internet connection", err)
	}

	return result.IP, nil
}
