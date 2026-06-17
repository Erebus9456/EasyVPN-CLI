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
			return
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

		case strings.Contains(choice, "Export macOS WireGuard Config"):
			// For macOS, option 8/3 triggers the connect flow which calls DarwinAdapter.CreateTunnel
			connectCmd.Run(connectCmd, []string{})

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
	index, _ := ui.ShowEnvOnboarding()
	if index == 3 {
		fmt.Println("\nPaste this into a file named .env in the project root:")
		fmt.Println("EASYVPN_API_TOKEN=\nEASYVPN_SUPABASE_URL=\nEASYVPN_SUPABASE_ANON_KEY=")
		os.Exit(0)
	}
	// Note: In a full 'beast-mode' build, Option 2 would loop
	// through PromptForString and write the .env file here.
	fmt.Println("Please create your .env file and restart the app.")
	os.Exit(0)
}

func handleError(err error) {
	if e, ok := err.(*models.EasyVPNError); ok {
		fmt.Printf("\n❌ [%s] %s\n", e.Code, e.Message)
		if e.Remediation != "" {
			fmt.Printf("💡 Fix: %s\n", e.Remediation)
		}
	} else {
		fmt.Printf("\n❌ Error: %v\n", err)
	}
	os.Exit(1)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
