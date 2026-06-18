package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Erebus9456/EasyVPN-CLI/internal/api"
	"github.com/Erebus9456/EasyVPN-CLI/internal/config"
	"github.com/Erebus9456/EasyVPN-CLI/internal/core"
	"github.com/Erebus9456/EasyVPN-CLI/internal/state"
	"github.com/Erebus9456/EasyVPN-CLI/internal/ui"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/models"
	"github.com/Erebus9456/EasyVPN-CLI/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	cfg        *config.Config
	vpnManager *core.Manager
	rootCmd    = &cobra.Command{
		Use:   "easyvpn",
		Short: "EasyVPN - High-performance VPN CLI",
		Run: func(cmd *cobra.Command, args []string) {
			runInteractiveMenu()
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(disconnectCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(ipCmd)

}

func initConfig() {
	var err error
	cfg, err = config.Load()

	// If config fails to load (missing .env), trigger onboarding
	if err != nil || cfg.ApiToken == "" {
		handleOnboarding()
		// Re-load after onboarding
		cfg, _ = config.Load()
	}

	utils.Initialize(cfg.LogLevel, []string{cfg.ApiToken, cfg.SupabaseKey})

	if err := cfg.EnsureConfigDir(); err != nil {
		handleError(err)
	}

	vpnManager, err = core.NewManager(cfg)
	if err != nil {
		handleError(err)
	}
}

// --- COMMANDS ---

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Verify and install system requirements",
	Run: func(cmd *cobra.Command, args []string) {
		adapter, _ := vpnManager.GetAdapter() // Internal helper
		ui.StartSpinner("Checking system requirements...")
		errs := adapter.CheckRequirements()
		if len(errs) > 0 {
			ui.StopSpinnerFail("Requirements check failed")
			for _, e := range errs {
				fmt.Printf("❌ %s (Fix: %s)\n", e.Message, e.Remediation)
			}

			fmt.Println("\n🛠️ Attempting automatic installation of missing dependencies...")
			if err := adapter.InstallDependencies(); err != nil {
				handleError(err)
				return
			}

			ui.StartSpinner("Re-checking system requirements...")
			errs = adapter.CheckRequirements()
			if len(errs) > 0 {
				ui.StopSpinnerFail("Requirements check failed after installation")
				for _, e := range errs {
					fmt.Printf("❌ %s (Fix: %s)\n", e.Message, e.Remediation)
				}
				return
			}
		}
		ui.StopSpinnerSuccess("System is ready for EasyVPN!")
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Select a server and connect",
	Run: func(cmd *cobra.Command, args []string) {
		ui.StartSpinner("Fetching server list...")
		discovery, _ := api.NewDiscoveryClient(cfg)
		servers, err := discovery.FetchActiveServers("")
		if err != nil {
			ui.StopSpinnerFail("Discovery failed")
			handleError(err)
		}
		ui.StopSpinnerSuccess(fmt.Sprintf("Found %d active servers", len(servers)))

		selected, err := ui.ShowServerSelection(servers)
		if err != nil {
			return
		}

		hostname, _ := os.Hostname()
		err = vpnManager.Connect(selected, hostname, true) // Enable kill-switch by default
		if err != nil {
			handleError(err)
		}
	},
}

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Stop the VPN tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		err := vpnManager.Disconnect()
		if err != nil {
			handleError(err)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current VPN reality",
	Run: func(cmd *cobra.Command, args []string) {
		store := state.NewStore(cfg.ConfigDir)
		adapter, _ := vpnManager.GetAdapter()
		reconciler := core.NewReconciler(store, adapter)

		ui.StartSpinner("Syncing with OS kernel...")
		currentState, err := reconciler.Sync()
		if err != nil {
			handleError(err)
		}
		ui.StopSpinnerSuccess("Status check complete")

		headers := []string{"Property", "Value"}
		data := [][]string{
			{"Connected", fmt.Sprintf("%v", currentState.IsConnected)},
			{"Server", currentState.ServerName},
			{"Interface", currentState.InterfaceName},
			{"Kill-Switch", fmt.Sprintf("%v", currentState.KillSwitch)},
		}
		ui.PrintTable(headers, data)
	},
}
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Display your current public IP address",
	Run: func(cmd *cobra.Command, args []string) {
		ui.StartSpinner("Querying public IP...")
		client := api.NewIPClient(cfg)
		ip, err := client.GetPublicIP()
		if err != nil {
			ui.StopSpinnerFail("Query failed")
			handleError(err)
		}
		ui.StopSpinnerSuccess(fmt.Sprintf("Your Public IP: %s", ip))
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a WireGuard config for mobile or desktop clients",
	Run: func(cmd *cobra.Command, args []string) {
		ui.StartSpinner("Fetching server list...")
		discovery, _ := api.NewDiscoveryClient(cfg)
		servers, err := discovery.FetchActiveServers("")
		if err != nil {
			ui.StopSpinnerFail("Discovery failed")
			handleError(err)
		}
		ui.StopSpinnerSuccess(fmt.Sprintf("Found %d active servers", len(servers)))

		selected, err := ui.ShowServerSelection(servers)
		if err != nil {
			return
		}

		exportPath, err := ui.PromptForExportPath()
		if err != nil {
			fmt.Println("Export canceled.")
			return
		}

		hostname, _ := os.Hostname()
		path, err := vpnManager.ExportConfig(selected, hostname, exportPath)
		if err != nil {
			printError(err)
			return
		}
		printExportInstructions(path)
	},
}

// --- LOGIC HELPERS ---
func runInteractiveMenu() {
	for {
		choice, err := ui.ShowMainMenu()
		if err != nil || strings.Contains(choice, "0) Exit") {
			fmt.Println("Goodbye!")
			return
		}

		// Use strings.Contains to map the menu text to the correct logic
		// regardless of what the leading number is.
		switch {
		case strings.Contains(choice, "Install all requirements"):
			setupCmd.Run(setupCmd, []string{})

		case strings.Contains(choice, "Check Current Public IP"):
			ipCmd.Run(ipCmd, []string{})

		case strings.Contains(choice, "Export WireGuard Config"):
			exportCmd.Run(exportCmd, []string{})

		case strings.Contains(choice, "List VPN Servers"):
			connectCmd.Run(connectCmd, []string{})

		case strings.Contains(choice, "Connect to VPN"):
			connectCmd.Run(connectCmd, []string{})

		case strings.Contains(choice, "Disconnect VPN"):
			disconnectCmd.Run(disconnectCmd, []string{})

		case strings.Contains(choice, "Show Connection Status"):
			statusCmd.Run(statusCmd, []string{})

		case strings.Contains(choice, "Kill-switch Management"):
			fmt.Println("Feature not available on this platform.")
		}

		fmt.Println("\n--- Press Enter to return to menu ---")
		fmt.Scanln()
	}
}
func handleOnboarding() {
	index, err := ui.ShowEnvOnboarding()
	if err != nil {
		fmt.Println("\nConfiguration canceled. Exiting.")
		os.Exit(1)
	}

	switch index {
	case 1:
		for {
			path, err := ui.PromptForString("Path to existing .env file", true)
			if err != nil {
				fmt.Println("\nConfiguration canceled. Exiting.")
				os.Exit(1)
			}
			if strings.HasPrefix(path, "~") {
				home, err := os.UserHomeDir()
				if err == nil {
					path = strings.Replace(path, "~", home, 1)
				}
			}
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("❌ Failed to read file at %s: %v. Please try again.\n", path, err)
				continue
			}
			err = os.WriteFile(".env", data, 0600)
			if err != nil {
				fmt.Printf("❌ Failed to create .env in project root: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Successfully loaded and copied configuration to project root .env.")
			break
		}

	case 2:
		apiToken, err := ui.PromptForConfigValue(
			"EASYVPN_API_TOKEN",
			"Your EasyVPN API Token. Used to authenticate with the EasyVPN API for fetching and managing server connections.",
			true,
			nil,
		)
		if err != nil {
			fmt.Println("\nConfiguration canceled. Exiting.")
			os.Exit(1)
		}

		supabaseUrl, err := ui.PromptForConfigValue(
			"EASYVPN_SUPABASE_URL",
			"Your Supabase Project URL. The endpoint where your database and API services are hosted (e.g., https://xyz.supabase.co).",
			true,
			func(val string) error {
				return utils.NewValidator().IsValidURL(val)
			},
		)
		if err != nil {
			fmt.Println("\nConfiguration canceled. Exiting.")
			os.Exit(1)
		}

		supabaseAnonKey, err := ui.PromptForConfigValue(
			"EASYVPN_SUPABASE_ANON_KEY",
			"Your Supabase Anonymous Key. Used for anonymous read access to fetch server endpoints and configurations.",
			true,
			nil,
		)
		if err != nil {
			fmt.Println("\nConfiguration canceled. Exiting.")
			os.Exit(1)
		}

		envContent := fmt.Sprintf(
			"EASYVPN_API_TOKEN=%s\nEASYVPN_SUPABASE_URL=%s\nEASYVPN_SUPABASE_ANON_KEY=%s\n",
			apiToken,
			supabaseUrl,
			supabaseAnonKey,
		)

		err = os.WriteFile(".env", []byte(envContent), 0600)
		if err != nil {
			fmt.Printf("❌ Failed to write .env file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Successfully created .env file in project root.")

	case 3:
		fmt.Println("\nPaste this into a file named .env in the project root:")
		fmt.Println("EASYVPN_API_TOKEN=\nEASYVPN_SUPABASE_URL=\nEASYVPN_SUPABASE_ANON_KEY=")
		os.Exit(0)

	default:
		fmt.Println("\nInvalid option. Exiting.")
		os.Exit(1)
	}
}

func printExportInstructions(path string) {
	fmt.Println("\n✅ WireGuard config exported successfully!")
	fmt.Printf("📂 Config saved to: %s\n", path)
	fmt.Println("--------------------------------------------------")
	fmt.Println("How to use:")
	fmt.Println("• macOS / iOS: Import the file in the WireGuard app")
	fmt.Println("• Android: Import from file or scan QR in the WireGuard app")
	fmt.Println("--------------------------------------------------")
}

func printError(err error) {
	if e, ok := err.(*models.EasyVPNError); ok {
		fmt.Printf("\n❌ [%s] %s\n", e.Code, e.Message)
		if e.Remediation != "" {
			fmt.Printf("💡 Fix: %s\n", e.Remediation)
		}
	} else {
		fmt.Printf("\n❌ Error: %v\n", err)
	}
}

func handleError(err error) {
	printError(err)
	os.Exit(1)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
