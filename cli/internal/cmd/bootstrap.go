package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fregataa/aami/cli/internal/client"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
)

var bootstrapTokenCmd = &cobra.Command{
	Use:     "bootstrap-token",
	Aliases: []string{"bt", "token"},
	Short:   "Manage bootstrap tokens",
	Long:    `Create, list, get, update, delete, validate, and use bootstrap tokens for node registration.`,
}

var bootstrapTokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new bootstrap token",
	Example: `  # Create a token with 7 days expiry
  aami bootstrap-token create --name=prod-token --max-uses=10 --expires=7d

  # Create a token with specific expiry date
  aami bootstrap-token create --name=staging-token --max-uses=50 --expires=2025-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		maxUses, _ := cmd.Flags().GetInt("max-uses")
		expiresStr, _ := cmd.Flags().GetString("expires")

		// Parse expiry
		var expiresAt time.Time
		var err error

		if expiresStr == "" {
			// Default: 30 days from now
			expiresAt = time.Now().Add(30 * 24 * time.Hour)
		} else if len(expiresStr) > 0 && expiresStr[len(expiresStr)-1] == 'd' {
			// Parse duration like "7d", "30d"
			days := 0
			_, err = fmt.Sscanf(expiresStr, "%dd", &days)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s (use format like '7d')", expiresStr)
			}
			expiresAt = time.Now().Add(time.Duration(days) * 24 * time.Hour)
		} else {
			// Parse as date
			expiresAt, err = time.Parse("2006-01-02", expiresStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %s (use format YYYY-MM-DD)", expiresStr)
			}
		}

		req := client.CreateBootstrapTokenRequest{
			Name:      name,
			MaxUses:   maxUses,
			ExpiresAt: expiresAt,
		}

		token, err := apiClient.CreateBootstrapToken(req)
		if err != nil {
			return fmt.Errorf("failed to create bootstrap token: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Created bootstrap token: %s", token.Name))
		output.PrintInfo(fmt.Sprintf("Token: %s", token.Token))
		output.PrintInfo(fmt.Sprintf("Save this token - it cannot be retrieved later!"))
		return getFormatter().Format(token)
	},
}

var bootstrapTokenListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List bootstrap tokens",
	Example: `  # List all bootstrap tokens
  aami bootstrap-token list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, err := apiClient.ListBootstrapTokens()
		if err != nil {
			return fmt.Errorf("failed to list bootstrap tokens: %w", err)
		}

		return getFormatter().Format(tokens)
	},
}

var bootstrapTokenGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a bootstrap token by ID",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get a bootstrap token
  aami bootstrap-token get 770e8400-e29b-41d4-a716-446655440007`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := apiClient.GetBootstrapToken(args[0])
		if err != nil {
			return fmt.Errorf("failed to get bootstrap token: %w", err)
		}

		return getFormatter().Format(token)
	},
}

var bootstrapTokenUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a bootstrap token",
	Args:  cobra.ExactArgs(1),
	Example: `  # Update max uses
  aami bootstrap-token update 770e8400-e29b-41d4-a716-446655440007 --max-uses=20

  # Extend expiry
  aami bootstrap-token update 770e8400-e29b-41d4-a716-446655440007 --expires=30d`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		req := client.UpdateBootstrapTokenRequest{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			req.Name = &name
		}
		if cmd.Flags().Changed("max-uses") {
			maxUses, _ := cmd.Flags().GetInt("max-uses")
			req.MaxUses = &maxUses
		}
		if cmd.Flags().Changed("expires") {
			expiresStr, _ := cmd.Flags().GetString("expires")
			var expiresAt time.Time
			var err error

			if len(expiresStr) > 0 && expiresStr[len(expiresStr)-1] == 'd' {
				days := 0
				_, err = fmt.Sscanf(expiresStr, "%dd", &days)
				if err != nil {
					return fmt.Errorf("invalid duration format: %s", expiresStr)
				}
				expiresAt = time.Now().Add(time.Duration(days) * 24 * time.Hour)
			} else {
				expiresAt, err = time.Parse("2006-01-02", expiresStr)
				if err != nil {
					return fmt.Errorf("invalid date format: %s", expiresStr)
				}
			}
			req.ExpiresAt = &expiresAt
		}

		token, err := apiClient.UpdateBootstrapToken(id, req)
		if err != nil {
			return fmt.Errorf("failed to update bootstrap token: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Updated bootstrap token: %s", token.Name))
		return getFormatter().Format(token)
	},
}

var bootstrapTokenDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a bootstrap token",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a bootstrap token
  aami bootstrap-token delete 770e8400-e29b-41d4-a716-446655440007`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if err := apiClient.DeleteBootstrapToken(id); err != nil {
			return fmt.Errorf("failed to delete bootstrap token: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Deleted bootstrap token: %s", id))
		return nil
	},
}

var bootstrapTokenValidateCmd = &cobra.Command{
	Use:   "validate <token>",
	Short: "Validate a bootstrap token",
	Args:  cobra.ExactArgs(1),
	Example: `  # Validate a token
  aami bootstrap-token validate abc123xyz...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenStr := args[0]

		token, err := apiClient.ValidateBootstrapToken(tokenStr)
		if err != nil {
			return fmt.Errorf("token validation failed: %w", err)
		}

		if token.IsValid {
			output.PrintSuccess("Token is valid")
		} else {
			output.PrintError("Token is invalid or expired")
		}

		return getFormatter().Format(token)
	},
}

var bootstrapTokenRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new node using bootstrap token",
	Example: `  # Register node with token
  aami bootstrap-token register \
    --token=abc123xyz... \
    --hostname=$(hostname) \
    --ip=$(hostname -I | awk '{print $1}')

  # Register with specific group
  aami bootstrap-token register \
    --token=abc123xyz... \
    --hostname=web-01 \
    --ip=10.0.1.100 \
    --group=<group-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		hostname, _ := cmd.Flags().GetString("hostname")
		ipAddress, _ := cmd.Flags().GetString("ip")
		groupID, _ := cmd.Flags().GetString("group")

		// Default hostname to OS hostname if not provided
		if hostname == "" {
			osHostname, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("failed to get hostname: %w", err)
			}
			hostname = osHostname
		}

		req := client.BootstrapRegisterRequest{
			Token:     token,
			Hostname:  hostname,
			IPAddress: ipAddress,
			GroupID:   groupID,
		}

		response, err := apiClient.RegisterNodeWithToken(req)
		if err != nil {
			return fmt.Errorf("node registration failed: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Successfully registered node: %s", hostname))
		output.PrintInfo(fmt.Sprintf("Target ID: %s", response.Target.ID))
		output.PrintInfo(fmt.Sprintf("Token usage: %d/%d", response.TokenUsage, response.TokenUsage+response.RemainingUses))

		return getFormatter().Format(response.Target)
	},
}

func init() {
	// Create command flags
	bootstrapTokenCreateCmd.Flags().StringP("name", "n", "", "Token name (required)")
	bootstrapTokenCreateCmd.Flags().Int("max-uses", 10, "Maximum number of uses")
	bootstrapTokenCreateCmd.Flags().String("expires", "30d", "Expiry (e.g., 7d, 30d, or 2025-12-31)")
	bootstrapTokenCreateCmd.MarkFlagRequired("name")

	// Update command flags
	bootstrapTokenUpdateCmd.Flags().StringP("name", "n", "", "New name")
	bootstrapTokenUpdateCmd.Flags().Int("max-uses", 0, "New max uses")
	bootstrapTokenUpdateCmd.Flags().String("expires", "", "New expiry")

	// Register command flags
	bootstrapTokenRegisterCmd.Flags().String("token", "", "Bootstrap token (required)")
	bootstrapTokenRegisterCmd.Flags().String("hostname", "", "Hostname (default: OS hostname)")
	bootstrapTokenRegisterCmd.Flags().String("ip", "", "IP address (required)")
	bootstrapTokenRegisterCmd.Flags().String("group", "", "Group ID (optional)")
	bootstrapTokenRegisterCmd.MarkFlagRequired("token")
	bootstrapTokenRegisterCmd.MarkFlagRequired("ip")

	// Add subcommands
	bootstrapTokenCmd.AddCommand(bootstrapTokenCreateCmd)
	bootstrapTokenCmd.AddCommand(bootstrapTokenListCmd)
	bootstrapTokenCmd.AddCommand(bootstrapTokenGetCmd)
	bootstrapTokenCmd.AddCommand(bootstrapTokenUpdateCmd)
	bootstrapTokenCmd.AddCommand(bootstrapTokenDeleteCmd)
	bootstrapTokenCmd.AddCommand(bootstrapTokenValidateCmd)
	bootstrapTokenCmd.AddCommand(bootstrapTokenRegisterCmd)
}
