package api

import (
	"encoding/json"
	"fmt"

	"github.com/Erebus9456/EasyVPN-CLI/internal/config"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/supabase-community/supabase-go"
)

// DiscoveryClient handles communication with Supabase
type DiscoveryClient struct {
	client *supabase.Client
}

// NewDiscoveryClient initializes a new Supabase client using the provided config
func NewDiscoveryClient(cfg *config.Config) (*DiscoveryClient, error) {
	client, err := supabase.NewClient(cfg.SupabaseUrl, cfg.SupabaseKey, nil)
	if err != nil {
		return nil, models.NewError(
			models.ErrDiscoveryFailed,
			"Failed to initialize Supabase client",
			"Check your EASYVPN_SUPABASE_URL and ANON_KEY",
			err,
		)
	}

	return &DiscoveryClient{
		client: client,
	}, err
}

// FetchActiveServers retrieves all servers from the registry where status is 'online'
func (d *DiscoveryClient) FetchActiveServers(region string) ([]models.Server, error) {
	// 1. Target the correct table: vpn_servers
	// 2. Filter for status 'online' (matches your SQL Enum)
	query := d.client.From("vpn_servers").
		Select("*", "1000", false).
		Eq("status", "online")

	// If a specific region is requested, add a filter
	if region != "" {
		query = query.Eq("region", region)
	}

	// Execute query
	data, _, err := query.Execute()
	if err != nil {
		// Beast-Mode: Log the raw error for debugging purposes
		fmt.Printf("DEBUG: Supabase Raw Error: %v\n", err)

		return nil, models.NewError(
			models.ErrDiscoveryFailed,
			"Failed to fetch servers from Supabase",
			"Check your internet connection or Supabase status",
			err,
		)
	}

	// Unmarshal the raw JSON data into our Server model slice
	var servers []models.Server
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, models.NewError(
			models.ErrInternal,
			"Failed to parse server data",
			"The database schema might have changed",
			err,
		)
	}

	// 3. Fallback logic for missing columns
	for i := range servers {
		// Since api_port isn't in your SQL yet, default it to 5000
		// so the AgentClient knows which port to call.
		if servers[i].APIPort == 0 {
			servers[i].APIPort = 5000
		}
	}

	return servers, nil
}
