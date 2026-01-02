package cli

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/xid"
)

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain error codes",
	Long:  "Provides detailed explanations for various error codes including NVIDIA Xid errors.",
}

var explainXidCmd = &cobra.Command{
	Use:   "xid [code]",
	Short: "Explain NVIDIA Xid error code",
	Long: `Explain an NVIDIA Xid error code with causes and recommended actions.

Xid errors are logged by the NVIDIA driver when GPU issues occur. This command
provides detailed information about what each error means and how to resolve it.

Examples:
  aami explain xid 79    # Explain Xid 79 (GPU fallen off bus)
  aami explain xid 48    # Explain Xid 48 (Double bit ECC error)`,
	Args: cobra.ExactArgs(1),
	RunE: runExplainXid,
}

var explainXidListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known Xid error codes",
	RunE:  runExplainXidList,
}

func init() {
	explainXidCmd.AddCommand(explainXidListCmd)
	explainCmd.AddCommand(explainXidCmd)
	rootCmd.AddCommand(explainCmd)
}

func runExplainXid(cmd *cobra.Command, args []string) error {
	code, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid Xid code: %s", args[0])
	}

	info, ok := xid.GetXidInfo(code)
	if !ok {
		return fmt.Errorf("unknown Xid code: %d\nRun 'aami explain xid list' to see all known codes", code)
	}

	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Header
	fmt.Println()
	fmt.Printf("┌─────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %s                             │\n", bold(fmt.Sprintf("Xid %d: %s", code, info.Name)))
	fmt.Printf("├─────────────────────────────────────────────────────────────────┤\n")

	// Severity
	severityText := info.Severity
	switch info.Severity {
	case "Critical":
		severityText = red(info.Severity)
	case "Warning":
		severityText = yellow(info.Severity)
	default:
		severityText = green(info.Severity)
	}
	fmt.Printf("│ Severity: %-53s │\n", severityText)
	fmt.Printf("│                                                                 │\n")

	// Description
	fmt.Printf("│ %s                                                       │\n", bold("Meaning:"))
	printWrapped(info.Description, 63, "│   ")
	fmt.Printf("│                                                                 │\n")

	// Causes
	fmt.Printf("│ %s                                                 │\n", bold("Common Causes:"))
	for i, cause := range info.Causes {
		fmt.Printf("│   %d. %-59s │\n", i+1, cause)
	}
	fmt.Printf("│                                                                 │\n")

	// Actions
	fmt.Printf("│ %s                                           │\n", bold("Recommended Actions:"))
	for i, action := range info.Actions {
		fmt.Printf("│   %d. %-59s │\n", i+1, action)
	}
	fmt.Printf("│                                                                 │\n")

	// Reference
	fmt.Printf("│ %s                                                     │\n", bold("Reference:"))
	fmt.Printf("│   %s│\n", cyan(fmt.Sprintf("%-63s", info.Reference)))
	fmt.Printf("└─────────────────────────────────────────────────────────────────┘\n")
	fmt.Println()

	return nil
}

func runExplainXidList(cmd *cobra.Command, args []string) error {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println("\nKnown Xid Error Codes:")
	fmt.Println()

	codes := xid.ListAllXids()
	sort.Ints(codes)

	for _, code := range codes {
		info, _ := xid.GetXidInfo(code)
		severity := info.Severity
		switch info.Severity {
		case "Critical":
			severity = red(fmt.Sprintf("%-8s", info.Severity))
		case "Warning":
			severity = yellow(fmt.Sprintf("%-8s", info.Severity))
		}
		fmt.Printf("  Xid %-3d  %s  %s\n", code, severity, info.Name)
	}
	fmt.Println()
	fmt.Println("Use 'aami explain xid <code>' for detailed information.")
	fmt.Println()

	return nil
}

func printWrapped(text string, width int, prefix string) {
	words := []rune(text)
	line := ""
	for _, word := range words {
		if len(line)+1 > width {
			fmt.Printf("%s%-63s │\n", prefix, line)
			line = string(word)
		} else {
			line += string(word)
		}
	}
	if line != "" {
		fmt.Printf("%s%-63s │\n", prefix, line)
	}
}
