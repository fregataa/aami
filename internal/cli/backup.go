package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/backup"
)

var (
	backupIncludeData bool
	backupOutputDir   string
	backupConfigOnly  bool
	backupForce       bool
	backupDryRun      bool
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore AAMI configuration",
	Long: `Backup and restore AAMI configuration and data.

Examples:
  aami backup create              # Create config backup
  aami backup create --include-data  # Include Prometheus/Grafana data
  aami backup list                # List available backups
  aami backup restore <file>      # Restore from backup
  aami backup restore <file> --config-only  # Restore config only`,
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new backup",
	Long: `Create a backup of AAMI configuration and optionally data.

Examples:
  aami backup create                    # Backup config files only
  aami backup create --include-data     # Include data directories
  aami backup create --output /tmp      # Save to custom directory`,
	RunE: runBackupCreate,
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	RunE:  runBackupList,
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore from a backup",
	Long: `Restore AAMI configuration and data from a backup file.

Examples:
  aami backup restore aami-backup-2024-01-01.tar.gz
  aami backup restore backup.tar.gz --config-only
  aami backup restore backup.tar.gz --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupRestore,
}

var backupVerifyCmd = &cobra.Command{
	Use:   "verify <backup-file>",
	Short: "Verify a backup file",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupVerify,
}

var backupContentsCmd = &cobra.Command{
	Use:   "contents <backup-file>",
	Short: "List contents of a backup file",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupContents,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Create subcommand
	backupCmd.AddCommand(backupCreateCmd)
	backupCreateCmd.Flags().BoolVar(&backupIncludeData, "include-data", false,
		"Include Prometheus/Grafana data in backup")
	backupCreateCmd.Flags().StringVarP(&backupOutputDir, "output", "o", "",
		"Output directory for backup file")

	// List subcommand
	backupCmd.AddCommand(backupListCmd)

	// Restore subcommand
	backupCmd.AddCommand(backupRestoreCmd)
	backupRestoreCmd.Flags().BoolVar(&backupConfigOnly, "config-only", false,
		"Only restore configuration files")
	backupRestoreCmd.Flags().BoolVar(&backupForce, "force", false,
		"Overwrite existing files without backup")
	backupRestoreCmd.Flags().BoolVar(&backupDryRun, "dry-run", false,
		"Show what would be restored without making changes")

	// Verify subcommand
	backupCmd.AddCommand(backupVerifyCmd)

	// Contents subcommand
	backupCmd.AddCommand(backupContentsCmd)
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	b := backup.NewBackup()

	opts := backup.DefaultBackupOptions()
	opts.IncludeData = backupIncludeData
	if backupOutputDir != "" {
		opts.OutputDir = backupOutputDir
	}

	fmt.Println("Creating backup...")

	result, err := b.Create(opts)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	color.Green("✓ Backup created successfully")
	fmt.Println()
	fmt.Printf("  File:       %s\n", result.FilePath)
	fmt.Printf("  Size:       %s\n", formatSize(result.Size))
	fmt.Printf("  Files:      %d\n", result.FileCount)
	fmt.Printf("  Data:       %v\n", result.IncludesData)
	fmt.Printf("  Created:    %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runBackupList(cmd *cobra.Command, args []string) error {
	b := backup.NewBackup()

	backups, err := b.List("")
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found")
		fmt.Println()
		fmt.Println("Create a backup with: aami backup create")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Size", "Created", "Type"})
	table.SetBorder(false)

	for _, bi := range backups {
		backupType := "config"
		if bi.IsFullBackup {
			backupType = "full"
		}

		table.Append([]string{
			bi.Name,
			bi.FormatSize(),
			bi.CreatedAt.Format("2006-01-02 15:04"),
			backupType,
		})
	}

	table.Render()
	return nil
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	backupPath := args[0]

	b := backup.NewBackup()

	// Verify backup first
	fmt.Println("Verifying backup...")
	verifyResult, err := b.Verify(backupPath)
	if err != nil {
		return fmt.Errorf("verify backup: %w", err)
	}

	if !verifyResult.IsValid {
		return fmt.Errorf("invalid backup: %s", verifyResult.Error)
	}

	fmt.Printf("Backup contains %d files\n", verifyResult.FileCount)
	fmt.Println()

	// Restore
	opts := backup.DefaultRestoreOptions()
	opts.ConfigOnly = backupConfigOnly
	opts.Force = backupForce
	opts.DryRun = backupDryRun

	if backupDryRun {
		fmt.Println("Dry run mode - no changes will be made")
		fmt.Println()
	}

	fmt.Println("Restoring backup...")

	result, err := b.Restore(backupPath, opts)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	if result.Success {
		color.Green("✓ Restore completed successfully")
	} else {
		color.Yellow("⚠ Restore completed with errors")
	}

	fmt.Println()
	fmt.Printf("  Files restored: %d\n", result.FilesRestored)
	fmt.Printf("  Files skipped:  %d\n", result.FilesSkipped)

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Println("Errors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	if !backupDryRun && result.FilesRestored > 0 {
		fmt.Println()
		fmt.Println("You may need to restart AAMI services for changes to take effect")
	}

	return nil
}

func runBackupVerify(cmd *cobra.Command, args []string) error {
	backupPath := args[0]

	b := backup.NewBackup()

	result, err := b.Verify(backupPath)
	if err != nil {
		return err
	}

	fmt.Printf("Backup: %s\n", backupPath)
	fmt.Printf("Size:   %s\n", formatSize(result.Size))
	fmt.Printf("Files:  %d\n", result.FileCount)
	fmt.Println()

	if result.IsValid {
		color.Green("✓ Backup is valid")

		fmt.Println()
		fmt.Printf("  Has metadata: %v\n", result.HasMetadata)
		fmt.Printf("  Has data:     %v\n", result.HasData)

		// Read and display metadata
		metadata, err := b.ReadMetadata(backupPath)
		if err == nil && len(metadata) > 0 {
			fmt.Println()
			fmt.Println("Metadata:")
			for k, v := range metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
	} else {
		color.Red("✗ Backup is invalid")
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}
	}

	return nil
}

func runBackupContents(cmd *cobra.Command, args []string) error {
	backupPath := args[0]

	b := backup.NewBackup()

	contents, err := b.ListContents(backupPath)
	if err != nil {
		return err
	}

	fmt.Printf("Contents of %s:\n\n", backupPath)
	for _, line := range contents {
		fmt.Println(line)
	}

	return nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
