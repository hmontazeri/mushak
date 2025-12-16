package cli

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/pkg/version"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update mushak to the latest version",
	Long:  `Update checks for the latest release on GitHub and updates the binary if a newer version is available.`,
	RunE:  runUpdate,
}

var updateCheck bool

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Check for updates without installing")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	currentVersion := version.GetVersion()

	ui.PrintInfo(fmt.Sprintf("Current version: %s", currentVersion))
	ui.PrintInfo("Checking for updates...")

	latest, found, err := selfupdate.DetectLatest("hmontazeri/mushak")
	if err != nil {
		return fmt.Errorf("failed to detect latest version: %w", err)
	}

	if !found {
		ui.PrintInfo("No release found")
		return nil
	}

	ui.PrintInfo(fmt.Sprintf("Latest version: %s", latest.Version))

	// If current version is dev, always show available version
	if currentVersion == "dev" {
		if updateCheck {
			ui.PrintSuccess(fmt.Sprintf("Version %s is available", latest.Version))
			ui.PrintInfo("Run 'mushak update' without --check to install")
			return nil
		}

		ui.PrintInfo("Note: You're running a development version")
		fmt.Printf("Would you like to install the latest release (%s)? (y/N): ", latest.Version)

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			ui.PrintInfo("Update cancelled")
			return nil
		}
	} else {
		// Check if update is available
		currentVer, err := semver.Make(currentVersion)
		if err != nil {
			return fmt.Errorf("failed to parse current version: %w", err)
		}
		if latest.Version.LTE(currentVer) {
			ui.PrintSuccess("You're already running the latest version")
			return nil
		}

		if updateCheck {
			ui.PrintSuccess(fmt.Sprintf("A new version is available: %s", latest.Version))
			ui.PrintInfo("Run 'mushak update' without --check to install")
			return nil
		}
	}

	ui.PrintInfo(fmt.Sprintf("Updating to %s...", latest.Version))

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Successfully updated to %s", latest.Version))
	ui.PrintInfo("Please restart mushak to use the new version")

	return nil
}
