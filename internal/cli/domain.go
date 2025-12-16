package cli

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/hooks"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/internal/utils"
	"github.com/spf13/cobra"
)

var domainCmd = &cobra.Command{
	Use:   "domain [NEW_DOMAIN]",
	Short: "Update the domain name for the deployed application",
	Long: `Update the domain name for the deployed application.
This command updates:
1. The local configuration
2. The running Caddy configuration on the server (immediate effect)
3. The deployment hook on the server (for future deployments)

Note: You must ensure your DNS records point to the server.`,
	Args: cobra.ExactArgs(1),
	RunE: runDomain,
}

var domainForce bool

func init() {
	rootCmd.AddCommand(domainCmd)
	domainCmd.Flags().BoolVarP(&domainForce, "force", "f", false, "Skip DNS confirmation")
}

func runDomain(cmd *cobra.Command, args []string) error {
	newDomain := args[0]
	if newDomain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Load deployment configuration
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Updating Domain")
	ui.PrintKeyValue("App", cfg.AppName)
	ui.PrintKeyValue("Old Domain", cfg.Domain)
	ui.PrintKeyValue("New Domain", newDomain)
	println()

	// Prompt for DNS confirmation
	if !domainForce {
		ui.PrintWarning("This will immediately route traffic for " + newDomain + " to this app.")
		ui.PrintInfo("Ensure your DNS records are updated to point to this server.")
		confirmed, err := utils.Confirm("Are your DNS records updated?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.PrintInfo("Aborted. Please update your DNS records first.")
			return nil
		}
		println()
	}

	// Connect to server
	ui.PrintInfo("Connecting to server...")
	client, err := ssh.NewClient(ssh.Config{
		Host: cfg.Host,
		User: cfg.User,
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer client.Close()
	ui.PrintSuccess("Connected to server")

	executor := ssh.NewExecutor(client)

	// Read existing Caddy config to get the current port
	caddyConfigPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", cfg.AppName)
	caddyConfig, err := executor.Run(fmt.Sprintf("cat %s", caddyConfigPath))
	if err != nil {
		return fmt.Errorf("failed to read existing Caddy config: %w", err)
	}

	port, err := getPortFromCaddyConfig(caddyConfig)
	if err != nil {
		return fmt.Errorf("failed to determine running port: %w", err)
	}

	// Update Caddy config
	if err := server.CreateAppCaddyConfig(executor, cfg.AppName, newDomain, port); err != nil {
		return fmt.Errorf("failed to update Caddy config: %w", err)
	}

	// Update post-receive hook
	ui.PrintInfo("Updating deployment hook...")
	hookScript := hooks.GeneratePostReceiveHook(cfg.AppName, newDomain, cfg.Branch)
	if err := server.InstallPostReceiveHook(executor, cfg.AppName, hookScript); err != nil {
		return fmt.Errorf("failed to install post-receive hook: %w", err)
	}

	// Update local config
	cfg.Domain = newDomain
	if err := config.SaveDeployConfig(cfg); err != nil {
		return fmt.Errorf("failed to update local config: %w", err)
	}
	ui.PrintSuccess("Local configuration updated")

	println()
	ui.PrintSuccess(fmt.Sprintf("Domain successfully changed to %s", newDomain))
	ui.PrintInfo("Make sure your DNS records are updated to point to this server.")

	return nil
}

// getPortFromCaddyConfig extracts the port number from a Caddy config block
// Expected format: reverse_proxy localhost:<PORT>
func getPortFromCaddyConfig(config string) (int, error) {
	re := regexp.MustCompile(`reverse_proxy\s+localhost:(\d+)`)
	matches := re.FindStringSubmatch(config)
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not find port in Caddy config")
	}

	port, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", matches[1])
	}

	return port, nil
}
