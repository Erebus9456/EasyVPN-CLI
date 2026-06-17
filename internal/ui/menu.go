package ui

import (
	"fmt"

	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
)

// ShowMainMenu presents the primary numeric options to the user
// internal/ui/menu.go

func ShowMainMenu() (string, error) {
	var options []string

	if runtime.GOOS == "darwin" {
		// Clean sequential numbering for macOS
		options = []string{
			"1) Install all requirements (wireguard-tools)",
			"2) Check Current Public IP",
			"3) Export macOS WireGuard Config",
			"0) Exit",
		}
	} else {
		// Sequential numbering for Linux/Windows
		options = []string{
			"1) Install all requirements and prepare machine",
			"2) List VPN Servers",
			"3) Connect to VPN",
			"4) Disconnect VPN",
			"5) Show Connection Status",
			"6) Kill-switch Management",
			"7) Check Current Public IP",
			"8) Export macOS WireGuard Config",
			"0) Exit",
		}
	}

	prompt := &survey.Select{
		Message: "EasyVPN Main Menu:",
		Options: options,
	}

	var result string
	err := survey.AskOne(prompt, &result)
	return result, err
}

// ShowServerSelection presents a list of available servers for the user to pick
func ShowServerSelection(servers []models.Server) (*models.Server, error) {
	if len(servers) == 0 {
		return nil, models.NewError(models.ErrDiscoveryFailed, "No servers available", "Ensure your Supabase configuration is correct", nil)
	}

	serverMap := make(map[string]models.Server)
	var options []string

	for _, s := range servers {
		label := fmt.Sprintf("[%s] %s (%s)", s.Region, s.Name, s.PublicIP)
		options = append(options, label)
		serverMap[label] = s
	}

	prompt := &survey.Select{
		Message: "Select a VPN Server:",
		Options: options,
	}

	var choice string
	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return nil, err
	}

	selected := serverMap[choice]
	return &selected, nil
}

// ShowEnvOnboarding handles the spec-mandated onboarding when .env is missing
func ShowEnvOnboarding() (int, error) {
	fmt.Println("\n⚠️  Configuration Missing (.env not found)")
	fmt.Println("To use EasyVPN, you need to provide your Supabase and API credentials.")

	options := []string{
		"1) Provide path to existing .env file",
		"2) Enter values individually and create .env in project root",
		"3) Print .env template and exit",
	}

	prompt := &survey.Select{
		Message: "How would you like to proceed?",
		Options: options,
	}

	var choice string
	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return 0, err
	}

	// Extract the number from the choice string (e.g., "1)" -> 1)
	var index int
	fmt.Sscanf(choice, "%d", &index)
	return index, nil
}

// PromptForString is a helper for manual input during onboarding
func PromptForString(message string, required bool) (string, error) {
	var val string
	prompt := &survey.Input{
		Message: message,
	}

	validate := func(val interface{}) error {
		if str, ok := val.(string); ok && len(str) == 0 && required {
			return fmt.Errorf("this field is required")
		}
		return nil
	}

	err := survey.AskOne(prompt, &val, survey.WithValidator(validate))
	return val, err
}

// ConfirmAction is a simple Yes/No prompt
func ConfirmAction(message string) bool {
	result := false
	prompt := &survey.Confirm{
		Message: message,
		Default: true,
	}
	survey.AskOne(prompt, &result)
	return result
}
