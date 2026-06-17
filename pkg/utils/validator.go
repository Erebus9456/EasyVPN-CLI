package utils

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

// Validator provides methods to verify data integrity
type Validator struct{}

// NewValidator creates a fresh instance of our gatekeeper
func NewValidator() *Validator {
	return &Validator{}
}

// IsValidIP checks if a string is a valid IPv4 or IPv6 address (with or without CIDR)
func (v *Validator) IsValidIP(ip string) error {
	// Check if it's a CIDR (e.g., 10.0.0.1/32)
	if strings.Contains(ip, "/") {
		_, _, err := net.ParseCIDR(ip)
		if err != nil {
			return models.NewError(models.ErrInvalidInput, "Invalid CIDR format", "Expected format like 10.0.0.1/32", err)
		}
		return nil
	}

	// Check if it's a plain IP
	if net.ParseIP(ip) == nil {
		return models.NewError(models.ErrInvalidInput, "Invalid IP address format", "Please provide a valid IPv4 or IPv6 address", nil)
	}
	return nil
}

// IsValidURL ensures the string is a properly formatted URL with a scheme
func (v *Validator) IsValidURL(input string) error {
	u, err := url.ParseRequestURI(input)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return models.NewError(models.ErrInvalidInput, "Invalid URL format", "URL must include scheme (http/https) and host", err)
	}
	return nil
}

// IsValidPort checks if a port is within the valid networking range
func (v *Validator) IsValidPort(port int) error {
	if port < 1 || port > 65535 {
		return models.NewError(models.ErrInvalidInput, "Invalid port number", "Port must be between 1 and 65535", nil)
	}
	return nil
}

// CheckRequiredFields ensures no critical configuration strings are empty
func (v *Validator) CheckRequiredFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return models.NewError(
				models.ErrConfigMissing,
				fmt.Sprintf("Missing required field: %s", name),
				fmt.Sprintf("Please set the %s in your .env file or configuration", name),
				nil,
			)
		}
	}
	return nil
}
