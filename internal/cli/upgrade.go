package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/upgrade"
)

var (
	upgradeCheck    bool
	upgradeRollback bool
	upgradeForce    bool
	upgradeVersion  string
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade AAMI to the latest version",
	Long: `Upgrade AAMI CLI to the latest version from GitHub releases.

Examples:
  aami upgrade --check      # Check for available updates
  aami upgrade              # Upgrade to the latest version
  aami upgrade --rollback   # Rollback to the previous version
  aami upgrade --version v1.0.0  # Upgrade/downgrade to specific version`,
	RunE: runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().BoolVar(&upgradeCheck, "check", false,
		"Check for available updates without upgrading")
	upgradeCmd.Flags().BoolVar(&upgradeRollback, "rollback", false,
		"Rollback to the previous version")
	upgradeCmd.Flags().BoolVar(&upgradeForce, "force", false,
		"Force upgrade even if already at latest version")
	upgradeCmd.Flags().StringVar(&upgradeVersion, "version", "",
		"Upgrade to a specific version")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	checker := upgrade.NewChecker()
	upgrader := upgrade.NewUpgrader()

	currentVersion := Version

	// Handle rollback
	if upgradeRollback {
		return handleRollback(upgrader)
	}

	// Handle check only
	if upgradeCheck {
		return handleCheck(checker, currentVersion)
	}

	// Perform upgrade
	return handleUpgrade(upgrader, currentVersion)
}

func handleCheck(checker *upgrade.Checker, currentVersion string) error {
	fmt.Println("Checking for updates...")
	fmt.Println()

	result, err := checker.CheckForUpdate(currentVersion)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", result.LatestVersion)
	fmt.Println()

	if result.UpdateAvailable {
		color.Green("Update available!")
		fmt.Println()
		fmt.Println("Release notes:")
		fmt.Println(result.ReleaseNotes)
		fmt.Println()
		fmt.Println("Run 'aami upgrade' to install the update")
	} else {
		color.Green("You are running the latest version")
	}

	return nil
}

func handleUpgrade(upgrader *upgrade.Upgrader, currentVersion string) error {
	fmt.Println("Checking for updates...")

	result, err := upgrader.Upgrade(currentVersion)
	if err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	if result.Success {
		if result.NewVersion != "" {
			color.Green("✓ %s", result.Message)
			fmt.Printf("\nBackup saved to: %s\n", result.BackupPath)
			fmt.Println("\nPlease restart AAMI to use the new version")
		} else {
			color.Green("✓ %s", result.Message)
		}
	}

	return nil
}

func handleRollback(upgrader *upgrade.Upgrader) error {
	// List available versions
	versions, err := upgrader.ListAvailableRollbacks()
	if err != nil {
		return fmt.Errorf("list rollbacks: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no rollback versions available")
	}

	fmt.Println("Available versions for rollback:")
	for _, v := range versions {
		fmt.Printf("  - %s\n", v)
	}
	fmt.Println()

	// Rollback to latest backup
	fmt.Printf("Rolling back to %s...\n", versions[0])

	if err := upgrader.RollbackTo(versions[0]); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	color.Green("✓ Successfully rolled back to %s", versions[0])
	fmt.Println("\nPlease restart AAMI to use the restored version")

	return nil
}

// Add releases subcommand
var releasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "List available releases",
	Long:  "List recent releases from GitHub",
	RunE:  runReleases,
}

func init() {
	upgradeCmd.AddCommand(releasesCmd)
}

func runReleases(cmd *cobra.Command, args []string) error {
	checker := upgrade.NewChecker()

	releases, err := checker.GetReleases(10)
	if err != nil {
		return fmt.Errorf("fetch releases: %w", err)
	}

	if len(releases) == 0 {
		fmt.Println("No releases found")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Version", "Published", "Type", "Assets"})
	table.SetBorder(false)

	for _, r := range releases {
		releaseType := "stable"
		if r.Prerelease {
			releaseType = "prerelease"
		}
		if r.Draft {
			releaseType = "draft"
		}

		table.Append([]string{
			r.TagName,
			r.PublishedAt.Format("2006-01-02"),
			releaseType,
			fmt.Sprintf("%d", len(r.Assets)),
		})
	}

	table.Render()
	return nil
}
