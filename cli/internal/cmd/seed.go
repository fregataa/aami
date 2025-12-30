package cmd

import (
	"fmt"

	"github.com/fregataa/aami/cli/internal/client"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	seedForce  bool
	seedDryRun bool
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed default templates into the database",
	Long: `Load default alert templates and script templates into the Config Server database.

This command calls the Config Server API to seed default templates.
The templates are read from the server's configured default files.`,
	Example: `  # Seed default templates (skip existing)
  aami seed

  # Seed and update existing templates
  aami seed --force

  # Dry run to see what would be seeded
  aami seed --dry-run`,
	RunE: runSeed,
}

func init() {
	// Seed command flags
	seedCmd.Flags().BoolVar(&seedForce, "force", false, "Update existing templates if they exist")
	seedCmd.Flags().BoolVar(&seedDryRun, "dry-run", false, "Show what would be seeded without making changes")
}

func runSeed(cmd *cobra.Command, args []string) error {
	req := client.SeedRequest{
		Force:  seedForce,
		DryRun: seedDryRun,
	}

	response, err := apiClient.Seed(req)
	if err != nil {
		return fmt.Errorf("failed to seed templates: %w", err)
	}

	// Display results
	if seedDryRun {
		output.PrintInfo("Dry run - no changes made")
	}

	fmt.Println()
	fmt.Println("Alert Templates:")
	fmt.Printf("  Created: %d\n", response.AlertTemplates.Created)
	fmt.Printf("  Updated: %d\n", response.AlertTemplates.Updated)
	fmt.Printf("  Skipped: %d\n", response.AlertTemplates.Skipped)

	fmt.Println()
	fmt.Println("Script Templates:")
	fmt.Printf("  Created: %d\n", response.ScriptTemplates.Created)
	fmt.Printf("  Updated: %d\n", response.ScriptTemplates.Updated)
	fmt.Printf("  Skipped: %d\n", response.ScriptTemplates.Skipped)

	if len(response.Errors) > 0 {
		fmt.Println()
		output.PrintError("Errors encountered:")
		for _, e := range response.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	if !seedDryRun && len(response.Errors) == 0 {
		totalCreated := response.AlertTemplates.Created + response.ScriptTemplates.Created
		totalUpdated := response.AlertTemplates.Updated + response.ScriptTemplates.Updated
		if totalCreated > 0 || totalUpdated > 0 {
			fmt.Println()
			output.PrintSuccess("Seed completed successfully")
		} else {
			fmt.Println()
			output.PrintInfo("No new templates to seed")
		}
	}

	return nil
}
