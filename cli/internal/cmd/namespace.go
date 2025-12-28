package cmd

import (
	"fmt"

	"github.com/fregataa/aami/cli/internal/client"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:     "namespace",
	Aliases: []string{"ns"},
	Short:   "Manage namespaces",
	Long:    `Create, list, get, update, and delete namespaces.`,
}

var namespaceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new namespace",
	Example: `  # Create a namespace
  aami namespace create --name=production --description="Production environment" --priority=100

  # Create with merge strategy
  aami namespace create --name=staging --priority=90 --merge-strategy=override`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetInt("priority")
		mergeStrategy, _ := cmd.Flags().GetString("merge-strategy")

		req := client.CreateNamespaceRequest{
			Name:           name,
			Description:    description,
			PolicyPriority: priority,
			MergeStrategy:  mergeStrategy,
		}

		ns, err := apiClient.CreateNamespace(req)
		if err != nil {
			return fmt.Errorf("failed to create namespace: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Created namespace: %s (ID: %s)", ns.Name, ns.ID))
		return getFormatter().Format(ns)
	},
}

var namespaceListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all namespaces",
	Example: `  # List all namespaces
  aami namespace list

  # List in JSON format
  aami namespace list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespaces, err := apiClient.ListNamespaces()
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}

		return getFormatter().Format(namespaces)
	},
}

var namespaceGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a namespace by ID or name",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get by ID
  aami namespace get 550e8400-e29b-41d4-a716-446655440000

  # Get by name
  aami namespace get --name=production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nameFlag, _ := cmd.Flags().GetString("name")

		var ns *client.Namespace
		var err error

		if nameFlag != "" {
			ns, err = apiClient.GetNamespaceByName(nameFlag)
		} else {
			ns, err = apiClient.GetNamespace(args[0])
		}

		if err != nil {
			return fmt.Errorf("failed to get namespace: %w", err)
		}

		return getFormatter().Format(ns)
	},
}

var namespaceUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a namespace",
	Args:  cobra.ExactArgs(1),
	Example: `  # Update description
  aami namespace update 550e8400-e29b-41d4-a716-446655440000 --description="Updated desc"

  # Update priority
  aami namespace update 550e8400-e29b-41d4-a716-446655440000 --priority=200`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		req := client.UpdateNamespaceRequest{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			req.Name = &name
		}
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			req.Description = &description
		}
		if cmd.Flags().Changed("priority") {
			priority, _ := cmd.Flags().GetInt("priority")
			req.PolicyPriority = &priority
		}
		if cmd.Flags().Changed("merge-strategy") {
			mergeStrategy, _ := cmd.Flags().GetString("merge-strategy")
			req.MergeStrategy = &mergeStrategy
		}

		ns, err := apiClient.UpdateNamespace(id, req)
		if err != nil {
			return fmt.Errorf("failed to update namespace: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Updated namespace: %s", ns.Name))
		return getFormatter().Format(ns)
	},
}

var namespaceDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a namespace",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a namespace
  aami namespace delete 550e8400-e29b-41d4-a716-446655440000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if err := apiClient.DeleteNamespace(id); err != nil {
			return fmt.Errorf("failed to delete namespace: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Deleted namespace: %s", id))
		return nil
	},
}

func init() {
	// Create command flags
	namespaceCreateCmd.Flags().StringP("name", "n", "", "Namespace name (required)")
	namespaceCreateCmd.Flags().StringP("description", "d", "", "Description")
	namespaceCreateCmd.Flags().IntP("priority", "p", 100, "Policy priority")
	namespaceCreateCmd.Flags().String("merge-strategy", "merge", "Merge strategy (merge|override)")
	namespaceCreateCmd.MarkFlagRequired("name")

	// Get command flags
	namespaceGetCmd.Flags().String("name", "", "Get by name instead of ID")

	// Update command flags
	namespaceUpdateCmd.Flags().StringP("name", "n", "", "New name")
	namespaceUpdateCmd.Flags().StringP("description", "d", "", "New description")
	namespaceUpdateCmd.Flags().IntP("priority", "p", 0, "New priority")
	namespaceUpdateCmd.Flags().String("merge-strategy", "", "New merge strategy")

	// Add subcommands
	namespaceCmd.AddCommand(namespaceCreateCmd)
	namespaceCmd.AddCommand(namespaceListCmd)
	namespaceCmd.AddCommand(namespaceGetCmd)
	namespaceCmd.AddCommand(namespaceUpdateCmd)
	namespaceCmd.AddCommand(namespaceDeleteCmd)
}
