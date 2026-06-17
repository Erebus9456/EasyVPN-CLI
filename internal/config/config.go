package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/utils"
	"github.com/spf13/viper"
)

// Config holds the validated configuration for the application
type Config struct {
	ApiToken          string
	SupabaseUrl       string
	SupabaseKey       string
	NodeApiBaseUrl    string
	DefaultRegion     string
	PublicIpCheckUrl  string
	LogLevel          string
	AllowedIPsDefault string
	DnsDefault        string
	ConfigDir         string
}

// Load initializes the configuration from .env and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// 1. Setup Defaults
	home, _ := os.UserHomeDir()
	defaultConfigDir := filepath.Join(home, ".easyvpn")

	v.SetDefault("EASYVPN_PUBLIC_IP_CHECK_URL", "https://api.ipify.org?format=json")
	v.SetDefault("EASYVPN_LOG_LEVEL", "info")
	v.SetDefault("EASYVPN_ALLOWED_IPS_DEFAULT", "0.0.0.0/0,::/0")
	v.SetDefault("EASYVPN_DNS_DEFAULT", "1.1.1.1")
	v.SetDefault("EASYVPN_CONFIG_DIR", defaultConfigDir)

	// 2. Read from .env file (if it exists)
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")  // Project root
	v.AddConfigPath(home) // Home directory fallback

	if err := v.ReadInConfig(); err != nil {
		// It's okay if .env is missing, as long as ENV vars are set
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, models.NewError(models.ErrInternal, "Error reading .env file", "Check file permissions", err)
		}
	}

	// 3. Enable Environment Variable Overrides
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 4. Map to Struct
	cfg := &Config{
		ApiToken:          v.GetString("EASYVPN_API_TOKEN"),
		SupabaseUrl:       v.GetString("EASYVPN_SUPABASE_URL"),
		SupabaseKey:       v.GetString("EASYVPN_SUPABASE_ANON_KEY"),
		NodeApiBaseUrl:    v.GetString("EASYVPN_NODE_API_BASE_URL"),
		DefaultRegion:     v.GetString("EASYVPN_DEFAULT_REGION"),
		PublicIpCheckUrl:  v.GetString("EASYVPN_PUBLIC_IP_CHECK_URL"),
		LogLevel:          v.GetString("EASYVPN_LOG_LEVEL"),
		AllowedIPsDefault: v.GetString("EASYVPN_ALLOWED_IPS_DEFAULT"),
		DnsDefault:        v.GetString("EASYVPN_DNS_DEFAULT"),
		ConfigDir:         v.GetString("EASYVPN_CONFIG_DIR"),
	}

	return cfg, nil
}

// Validate ensures all mandatory fields are present
func (c *Config) Validate() error {
	validator := utils.NewValidator()

	required := map[string]string{
		"EASYVPN_API_TOKEN":         c.ApiToken,
		"EASYVPN_SUPABASE_URL":      c.SupabaseUrl,
		"EASYVPN_SUPABASE_ANON_KEY": c.SupabaseKey,
	}

	// Check for empty fields
	if err := validator.CheckRequiredFields(required); err != nil {
		return err
	}

	// Validate URL format
	if err := validator.IsValidURL(c.SupabaseUrl); err != nil {
		return models.NewError(models.ErrInvalidInput, "Invalid Supabase URL", "Check EASYVPN_SUPABASE_URL in .env", err)
	}

	return nil
}

// EnsureConfigDir creates the ~/.easyvpn directory if it doesn't exist
func (c *Config) EnsureConfigDir() error {
	if _, err := os.Stat(c.ConfigDir); os.IsNotExist(err) {
		err := os.MkdirAll(c.ConfigDir, 0700) // Restricted permissions
		if err != nil {
			return models.NewError(models.ErrInternal, "Failed to create config directory", "Check folder permissions", err)
		}
	}
	return nil
}
