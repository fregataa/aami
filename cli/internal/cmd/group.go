package cmd

import (
	"fmt"

	"github.com/fregataa/aami/cli/internal/client"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
	Long:  `Create, list, get, update, and delete groups.`,
}

var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group",
	Example: `  # Create a group
  aami group create --name=web-tier --namespace=<ns-id> --description="Web servers"

  # Create a child group
  aami group create --name=web-prod --namespace=<ns-id> --parent=<parent-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		namespaceID, _ := cmd.Flags().GetString("namespace")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetInt("priority")
		parentID, _ := cmd.Flags().GetString("parent")

		req := client.CreateGroupRequest{
			Name:        name,
			NamespaceID: namespaceID,
			Description: description,
			Priority:    priority,
		}

		if parentID != "" {
			req.ParentID = &parentID
		}

		group, err := apiClient.CreateGroup(req)
		if err != nil {
			return fmt.Errorf("failed to create group: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Created group: %s (ID: %s)", group.Name, group.ID))
		return getFormatter().Format(group)
	},
}

var groupListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List groups",
	Example: `  # List all groups
  aami group list

  # List groups in a namespace
  aami group list --namespace=<ns-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespaceID, _ := cmd.Flags().GetString("namespace")

		var groups []client.Group
		var err error

		if namespaceID != "" {
			groups, err = apiClient.ListGroupsByNamespace(namespaceID)
		} else {
			groups, err = apiClient.ListGroups()
		}

		if err != nil {
			return fmt.Errorf("failed to list groups: %w", err)
		}

		return getFormatter().Format(groups)
	},
}

var groupGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a group by ID",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get a group
  aami group get 660e8400-e29b-41d4-a716-446655440001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		group, err := apiClient.GetGroup(args[0])
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		return getFormatter().Format(group)
	},
}

var groupUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a group",
	Args:  cobra.ExactArgs(1),
	Example: `  # Update description
  aami group update 660e8400-e29b-41d4-a716-446655440001 --description="Updated"

  # Change parent
  aami group update 660e8400-e29b-41d4-a716-446655440001 --parent=<new-parent-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		req := client.UpdateGroupRequest{}

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
			req.Priority = &priority
		}
		if cmd.Flags().Changed("parent") {
			parentID, _ := cmd.Flags().GetString("parent")
			req.ParentID = &parentID
		}

		group, err := apiClient.UpdateGroup(id, req)
		if err != nil {
			return fmt.Errorf("failed to update group: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Updated group: %s", group.Name))
		return getFormatter().Format(group)
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a group
  aami group delete 660e8400-e29b-41d4-a716-446655440001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if err := apiClient.DeleteGroup(id); err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Deleted group: %s", id))
		return nil
	},
}

var groupChildrenCmd = &cobra.Command{
	Use:   "children <id>",
	Short: "List child groups",
	Args:  cobra.ExactArgs(1),
	Example: `  # List child groups
  aami group children 660e8400-e29b-41d4-a716-446655440001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groups, err := apiClient.GetGroupChildren(args[0])
		if err != nil {
			return fmt.Errorf("failed to get child groups: %w", err)
		}

		return getFormatter().Format(groups)
	},
}

var groupAncestorsCmd = &cobra.Command{
	Use:   "ancestors <id>",
	Short: "List ancestor groups",
	Args:  cobra.ExactArgs(1),
	Example: `  # List ancestor groups
  aami group ancestors 660e8400-e29b-41d4-a716-446655440001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groups, err := apiClient.GetGroupAncestors(args[0])
		if err != nil {
			return fmt.Errorf("failed to get ancestor groups: %w", err)
		}

		return getFormatter().Format(groups)
	},
}

func init() {
	// Create command flags
	groupCreateCmd.Flags().StringP("name", "n", "", "Group name (required)")
	groupCreateCmd.Flags().String("namespace", "", "Namespace ID (required)")
	groupCreateCmd.Flags().StringP("description", "d", "", "Description")
	groupCreateCmd.Flags().IntP("priority", "p", 100, "Priority")
	groupCreateCmd.Flags().String("parent", "", "Parent group ID")
	groupCreateCmd.MarkFlagRequired("name")
	groupCreateCmd.MarkFlagRequired("namespace")

	// List command flags
	groupListCmd.Flags().String("namespace", "", "Filter by namespace ID")

	// Update command flags
	groupUpdateCmd.Flags().StringP("name", "n", "", "New name")
	groupUpdateCmd.Flags().StringP("description", "d", "", "New description")
	groupUpdateCmd.Flags().IntP("priority", "p", 0, "New priority")
	groupUpdateCmd.Flags().String("parent", "", "New parent group ID")

	// Add subcommands
	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupGetCmd)
	groupCmd.AddCommand(groupUpdateCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	groupCmd.AddCommand(groupChildrenCmd)
	groupCmd.AddCommand(groupAncestorsCmd)
}
