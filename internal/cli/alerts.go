package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage alert rules",
	Long:  "List, apply, and manage Prometheus alert rules.",
}

var alertsListPresetsCmd = &cobra.Command{
	Use:   "list-presets",
	Short: "List available alert presets",
	RunE:  runAlertsListPresets,
}

var alertsApplyPresetCmd = &cobra.Command{
	Use:   "apply-preset [name]",
	Short: "Apply an alert preset",
	Long: `Apply a predefined alert preset to Prometheus.

Available presets:
  gpu-basic       Basic GPU monitoring (3 rules)
  gpu-production  Comprehensive GPU monitoring (8 rules)

Examples:
  aami alerts apply-preset gpu-production`,
	Args: cobra.ExactArgs(1),
	RunE: runAlertsApplyPreset,
}

var alertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active alert rules",
	RunE:  runAlertsList,
}

// Preset definitions
type alertPreset struct {
	Name        string
	Description string
	Rules       []alertRule
}

type alertRule struct {
	Name        string
	Expr        string
	For         string
	Severity    string
	Summary     string
	Description string
}

var presets = map[string]alertPreset{
	"gpu-basic": {
		Name:        "gpu-basic",
		Description: "Basic GPU monitoring alerts",
		Rules: []alertRule{
			{
				Name:        "GPUTemperatureCritical",
				Expr:        "DCGM_FI_DEV_GPU_TEMP > 85",
				For:         "5m",
				Severity:    "critical",
				Summary:     "GPU temperature critical on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} temperature is {{ $value }}°C",
			},
			{
				Name:        "GPUMemoryHigh",
				Expr:        "DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL * 100 > 95",
				For:         "10m",
				Severity:    "warning",
				Summary:     "GPU memory usage high on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} memory usage is {{ $value }}%",
			},
			{
				Name:        "NodeDown",
				Expr:        "up{job=\"node\"} == 0",
				For:         "1m",
				Severity:    "critical",
				Summary:     "Node {{ $labels.instance }} is down",
				Description: "Node exporter has been unreachable for more than 1 minute",
			},
		},
	},
	"gpu-production": {
		Name:        "gpu-production",
		Description: "Comprehensive GPU monitoring for production",
		Rules: []alertRule{
			{
				Name:        "GPUTemperatureCritical",
				Expr:        "DCGM_FI_DEV_GPU_TEMP > 85",
				For:         "5m",
				Severity:    "critical",
				Summary:     "GPU temperature critical on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} temperature is {{ $value }}°C",
			},
			{
				Name:        "GPUTemperatureWarning",
				Expr:        "DCGM_FI_DEV_GPU_TEMP > 75",
				For:         "10m",
				Severity:    "warning",
				Summary:     "GPU temperature warning on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} temperature is {{ $value }}°C",
			},
			{
				Name:        "GPUMemoryHigh",
				Expr:        "DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL * 100 > 95",
				For:         "10m",
				Severity:    "warning",
				Summary:     "GPU memory usage high on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} memory usage is {{ $value }}%",
			},
			{
				Name:        "GPUMemoryLeak",
				Expr:        "(DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL * 100 > 95) and (DCGM_FI_DEV_GPU_UTIL < 5)",
				For:         "30m",
				Severity:    "warning",
				Summary:     "Possible GPU memory leak on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} has high memory usage but low utilization",
			},
			{
				Name:        "GPUECCErrors",
				Expr:        "increase(DCGM_FI_DEV_ECC_DBE_VOL_TOTAL[24h]) > 100",
				For:         "0m",
				Severity:    "critical",
				Summary:     "High ECC error count on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} has reported over 100 ECC errors in 24h",
			},
			{
				Name:        "GPUXidError",
				Expr:        "increase(DCGM_FI_DEV_XID_ERRORS[5m]) > 0",
				For:         "0m",
				Severity:    "critical",
				Summary:     "Xid error on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} reported Xid error",
			},
			{
				Name:        "GPUNVLinkError",
				Expr:        "increase(DCGM_FI_DEV_NVLINK_CRC_FLIT_ERROR_COUNT_TOTAL[1h]) > 0",
				For:         "0m",
				Severity:    "warning",
				Summary:     "NVLink error on {{ $labels.instance }}",
				Description: "GPU {{ $labels.gpu }} has NVLink CRC errors",
			},
			{
				Name:        "NodeDown",
				Expr:        "up{job=\"node\"} == 0",
				For:         "1m",
				Severity:    "critical",
				Summary:     "Node {{ $labels.instance }} is down",
				Description: "Node exporter has been unreachable for more than 1 minute",
			},
		},
	},
}

func init() {
	alertsCmd.AddCommand(alertsListPresetsCmd)
	alertsCmd.AddCommand(alertsApplyPresetCmd)
	alertsCmd.AddCommand(alertsListCmd)
	rootCmd.AddCommand(alertsCmd)
}

func runAlertsListPresets(cmd *cobra.Command, args []string) error {
	fmt.Println("\nAvailable Alert Presets:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Preset", "Description", "Rules"})
	table.SetBorder(true)

	for name, preset := range presets {
		table.Append([]string{
			name,
			preset.Description,
			fmt.Sprintf("%d", len(preset.Rules)),
		})
	}

	table.Render()
	fmt.Println()
	fmt.Println("Apply a preset with: aami alerts apply-preset <name>")
	fmt.Println()

	return nil
}

func runAlertsApplyPreset(cmd *cobra.Command, args []string) error {
	presetName := args[0]
	preset, ok := presets[presetName]
	if !ok {
		return fmt.Errorf("unknown preset: %s\nRun 'aami alerts list-presets' to see available presets", presetName)
	}

	green := color.New(color.FgGreen).SprintFunc()

	// Generate Prometheus rules file
	rulesDir := "/etc/aami/rules"
	rulesFile := filepath.Join(rulesDir, fmt.Sprintf("%s.yaml", presetName))

	// Ensure directory exists
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("create rules directory: %w", err)
	}

	// Generate YAML content
	content := generatePrometheusRules(preset)

	if err := os.WriteFile(rulesFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write rules file: %w", err)
	}

	fmt.Printf("%s Applied preset %s (%d rules)\n", green("✓"), presetName, len(preset.Rules))
	fmt.Printf("  Rules file: %s\n", rulesFile)
	fmt.Println()
	fmt.Println("Note: Reload Prometheus to activate the rules:")
	fmt.Println("  curl -X POST http://localhost:9090/-/reload")
	fmt.Println()

	return nil
}

func runAlertsList(cmd *cobra.Command, args []string) error {
	rulesDir := "/etc/aami/rules"

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No alert rules configured.")
			fmt.Println("Apply a preset with: aami alerts apply-preset gpu-production")
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No alert rules configured.")
		fmt.Println("Apply a preset with: aami alerts apply-preset gpu-production")
		return nil
	}

	fmt.Println("\nActive Alert Rules:")
	fmt.Println()

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml") {
			name := strings.TrimSuffix(strings.TrimSuffix(entry.Name(), ".yaml"), ".yml")
			fmt.Printf("  • %s\n", name)
		}
	}

	fmt.Println()
	return nil
}

func generatePrometheusRules(preset alertPreset) string {
	var sb strings.Builder

	sb.WriteString("# Generated by AAMI - Do not edit manually\n")
	sb.WriteString(fmt.Sprintf("# Preset: %s\n\n", preset.Name))
	sb.WriteString("groups:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s\n", preset.Name))
	sb.WriteString("    rules:\n")

	for _, rule := range preset.Rules {
		sb.WriteString(fmt.Sprintf("      - alert: %s\n", rule.Name))
		sb.WriteString(fmt.Sprintf("        expr: %s\n", rule.Expr))
		if rule.For != "" && rule.For != "0m" {
			sb.WriteString(fmt.Sprintf("        for: %s\n", rule.For))
		}
		sb.WriteString("        labels:\n")
		sb.WriteString(fmt.Sprintf("          severity: %s\n", rule.Severity))
		sb.WriteString("        annotations:\n")
		sb.WriteString(fmt.Sprintf("          summary: \"%s\"\n", rule.Summary))
		sb.WriteString(fmt.Sprintf("          description: \"%s\"\n", rule.Description))
		sb.WriteString("\n")
	}

	return sb.String()
}
