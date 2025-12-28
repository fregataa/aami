package cmd

import (
	"fmt"
	"strings"

	"github.com/fregataa/aami/cli/internal/client"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
)

var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Manage targets",
	Long:  `Create, list, get, update, and delete targets (monitored nodes).`,
}

var targetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new target",
	Example: `  # Create a target
  aami target create --hostname=web-01 --ip=10.0.1.100 --group=<group-id>

  # Create with multiple groups
  aami target create --hostname=web-01 --ip=10.0.1.100 --groups=<id1>,<id2>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname, _ := cmd.Flags().GetString("hostname")
		ipAddress, _ := cmd.Flags().GetString("ip")
		groupID, _ := cmd.Flags().GetString("group")
		groupsStr, _ := cmd.Flags().GetString("groups")

		var groupIDs []string
		if groupsStr != "" {
			groupIDs = strings.Split(groupsStr, ",")
		} else if groupID != "" {
			groupIDs = []string{groupID}
		}

		req := client.CreateTargetRequest{
			Hostname:  hostname,
			IPAddress: ipAddress,
			GroupIDs:  groupIDs,
		}

		target, err := apiClient.CreateTarget(req)
		if err != nil {
			return fmt.Errorf("failed to create target: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Created target: %s (ID: %s)", target.Hostname, target.ID))
		return getFormatter().Format(target)
	},
}

var targetListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List targets",
	Example: `  # List all targets
  aami target list

  # List targets in a group
  aami target list --group=<group-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID, _ := cmd.Flags().GetString("group")

		var targets []client.Target
		var err error

		if groupID != "" {
			targets, err = apiClient.ListTargetsByGroup(groupID)
		} else {
			targets, err = apiClient.ListTargets()
		}

		if err != nil {
			return fmt.Errorf("failed to list targets: %w", err)
		}

		return getFormatter().Format(targets)
	},
}

var targetGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a target by ID or hostname",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get by ID
  aami target get 880e8400-e29b-41d4-a716-446655440003

  # Get by hostname
  aami target get --hostname=web-01`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hostnameFlag, _ := cmd.Flags().GetString("hostname")

		var target *client.Target
		var err error

		if hostnameFlag != "" {
			target, err = apiClient.GetTargetByHostname(hostnameFlag)
		} else {
			target, err = apiClient.GetTarget(args[0])
		}

		if err != nil {
			return fmt.Errorf("failed to get target: %w", err)
		}

		return getFormatter().Format(target)
	},
}

var targetUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a target",
	Args:  cobra.ExactArgs(1),
	Example: `  # Update hostname
  aami target update 880e8400-e29b-41d4-a716-446655440003 --hostname=web-01-new

  # Update groups
  aami target update 880e8400-e29b-41d4-a716-446655440003 --groups=<id1>,<id2>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		req := client.UpdateTargetRequest{}

		if cmd.Flags().Changed("hostname") {
			hostname, _ := cmd.Flags().GetString("hostname")
			req.Hostname = &hostname
		}
		if cmd.Flags().Changed("ip") {
			ipAddress, _ := cmd.Flags().GetString("ip")
			req.IPAddress = &ipAddress
		}
		if cmd.Flags().Changed("groups") {
			groupsStr, _ := cmd.Flags().GetString("groups")
			groupIDs := strings.Split(groupsStr, ",")
			req.GroupIDs = &groupIDs
		}

		target, err := apiClient.UpdateTarget(id, req)
		if err != nil {
			return fmt.Errorf("failed to update target: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Updated target: %s", target.Hostname))
		return getFormatter().Format(target)
	},
}

var targetDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a target",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a target
  aami target delete 880e8400-e29b-41d4-a716-446655440003`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if err := apiClient.DeleteTarget(id); err != nil {
			return fmt.Errorf("failed to delete target: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Deleted target: %s", id))
		return nil
	},
}

var targetStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Update target status",
	Args:  cobra.ExactArgs(1),
	Example: `  # Set target status to active
  aami target status 880e8400-e29b-41d4-a716-446655440003 --status=active

  # Set target status to inactive
  aami target status 880e8400-e29b-41d4-a716-446655440003 --status=inactive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		status, _ := cmd.Flags().GetString("status")

		if err := apiClient.UpdateTargetStatus(id, status); err != nil {
			return fmt.Errorf("failed to update target status: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Updated target status to: %s", status))
		return nil
	},
}

var targetHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat <id>",
	Short: "Send heartbeat for a target",
	Args:  cobra.ExactArgs(1),
	Example: `  # Send heartbeat
  aami target heartbeat 880e8400-e29b-41d4-a716-446655440003`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if err := apiClient.HeartbeatTarget(id); err != nil {
			return fmt.Errorf("failed to send heartbeat: %w", err)
		}

		output.PrintSuccess("Heartbeat sent successfully")
		return nil
	},
}

func init() {
	// Create command flags
	targetCreateCmd.Flags().String("hostname", "", "Target hostname (required)")
	targetCreateCmd.Flags().String("ip", "", "IP address (required)")
	targetCreateCmd.Flags().String("group", "", "Group ID")
	targetCreateCmd.Flags().String("groups", "", "Comma-separated group IDs")
	targetCreateCmd.MarkFlagRequired("hostname")
	targetCreateCmd.MarkFlagRequired("ip")

	// List command flags
	targetListCmd.Flags().String("group", "", "Filter by group ID")

	// Get command flags
	targetGetCmd.Flags().String("hostname", "", "Get by hostname instead of ID")

	// Update command flags
	targetUpdateCmd.Flags().String("hostname", "", "New hostname")
	targetUpdateCmd.Flags().String("ip", "", "New IP address")
	targetUpdateCmd.Flags().String("groups", "", "New comma-separated group IDs")

	// Status command flags
	targetStatusCmd.Flags().String("status", "", "Status (active|inactive)")
	targetStatusCmd.MarkFlagRequired("status")

	// Add subcommands
	targetCmd.AddCommand(targetCreateCmd)
	targetCmd.AddCommand(targetListCmd)
	targetCmd.AddCommand(targetGetCmd)
	targetCmd.AddCommand(targetUpdateCmd)
	targetCmd.AddCommand(targetDeleteCmd)
	targetCmd.AddCommand(targetStatusCmd)
	targetCmd.AddCommand(targetHeartbeatCmd)
}
