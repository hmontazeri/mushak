package cli

import (
	"fmt"
	"os"

	"github.com/blang/semver"
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

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Println("Checking for updates...")

	latest, found, err := selfupdate.DetectLatest("hmontazeri/mushak")
	if err != nil {
		return fmt.Errorf("failed to detect latest version: %w", err)
	}

	if !found {
		fmt.Println("No release found")
		return nil
	}

	fmt.Printf("Latest version: %s\n", latest.Version)

	// If current version is dev, always show available version
	if currentVersion == "dev" {
		if updateCheck {
			fmt.Printf("\nVersion %s is available\n", latest.Version)
			fmt.Println("Run 'mushak update' without --check to install")
			return nil
		}

		fmt.Println("\nNote: You're running a development version")
		fmt.Printf("Would you like to install the latest release (%s)? (y/N): ", latest.Version)

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Update cancelled")
			return nil
		}
	} else {
		// Check if update is available
		currentVer, err := semver.Make(currentVersion)
		if err != nil {
			return fmt.Errorf("failed to parse current version: %w", err)
		}
		if latest.Version.LTE(currentVer) {
			fmt.Println("✓ You're already running the latest version")
			return nil
		}

		if updateCheck {
			fmt.Printf("\nA new version is available: %s\n", latest.Version)
			fmt.Println("Run 'mushak update' without --check to install")
			return nil
		}
	}

	fmt.Printf("\nUpdating to %s...\n", latest.Version)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("✓ Successfully updated to %s\n", latest.Version)
	fmt.Println("\nPlease restart mushak to use the new version")

	return nil
}
