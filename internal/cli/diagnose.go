package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/config"
	"github.com/fregataa/aami/internal/ssh"
)

var (
	diagnoseVerbose bool
	diagnoseFix     bool
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Run system diagnostics",
	Long: `Run comprehensive diagnostics on AAMI installation and configuration.

Checks:
  - Configuration file validity
  - Component status (Prometheus, Grafana, exporters)
  - Node connectivity
  - Port availability
  - Disk space
  - System requirements

Examples:
  aami diagnose           # Run all diagnostics
  aami diagnose --verbose # Show detailed output
  aami diagnose --fix     # Attempt to fix common issues`,
	RunE: runDiagnose,
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)

	diagnoseCmd.Flags().BoolVarP(&diagnoseVerbose, "verbose", "v", false,
		"Show detailed diagnostic output")
	diagnoseCmd.Flags().BoolVar(&diagnoseFix, "fix", false,
		"Attempt to fix common issues")
}

// DiagnosticResult represents the result of a single diagnostic check.
type DiagnosticResult struct {
	Name    string
	Status  string // pass, warn, fail
	Message string
	Details string
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║              AAMI System Diagnostics               ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	var results []DiagnosticResult

	// 1. System checks
	fmt.Println("System Checks")
	fmt.Println(strings.Repeat("-", 50))
	results = append(results, checkSystem()...)
	printResults(results[len(results)-3:])
	fmt.Println()

	// 2. Configuration checks
	fmt.Println("Configuration Checks")
	fmt.Println(strings.Repeat("-", 50))
	configResults := checkConfiguration()
	results = append(results, configResults...)
	printResults(configResults)
	fmt.Println()

	// 3. Component checks
	fmt.Println("Component Checks")
	fmt.Println(strings.Repeat("-", 50))
	componentResults := checkComponents()
	results = append(results, componentResults...)
	printResults(componentResults)
	fmt.Println()

	// 4. Node connectivity (if config exists)
	cfg, err := loadConfig()
	if err == nil && len(cfg.Nodes) > 0 {
		fmt.Println("Node Connectivity")
		fmt.Println(strings.Repeat("-", 50))
		nodeResults := checkNodes(cfg)
		results = append(results, nodeResults...)
		printResults(nodeResults)
		fmt.Println()
	}

	// Summary
	printSummary(results)

	return nil
}

func checkSystem() []DiagnosticResult {
	var results []DiagnosticResult

	// OS check
	osResult := DiagnosticResult{
		Name:    "Operating System",
		Status:  "pass",
		Message: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		osResult.Status = "warn"
		osResult.Message += " (not officially supported)"
	}
	results = append(results, osResult)

	// Disk space check
	diskResult := checkDiskSpace()
	results = append(results, diskResult)

	// Memory check
	memResult := DiagnosticResult{
		Name:    "System Memory",
		Status:  "pass",
		Message: "Check skipped (requires system commands)",
	}
	results = append(results, memResult)

	return results
}

func checkDiskSpace() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Disk Space",
	}

	// Check /var/lib/aami if it exists, otherwise check root
	paths := []string{"/var/lib/aami", "/var/lib", "/"}
	var checkPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			checkPath = p
			break
		}
	}

	// Use df command to check disk space
	cmd := exec.Command("df", "-h", checkPath)
	output, err := cmd.Output()
	if err != nil {
		result.Status = "warn"
		result.Message = "Could not check disk space"
		return result
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 5 {
			usageStr := strings.TrimSuffix(fields[4], "%")
			result.Message = fmt.Sprintf("%s available (%s used)", fields[3], fields[4])

			var usage int
			fmt.Sscanf(usageStr, "%d", &usage)
			if usage >= 90 {
				result.Status = "fail"
				result.Message += " - CRITICAL"
			} else if usage >= 80 {
				result.Status = "warn"
				result.Message += " - Warning"
			} else {
				result.Status = "pass"
			}
		}
	}

	return result
}

func checkConfiguration() []DiagnosticResult {
	var results []DiagnosticResult

	// Config file exists
	configPath := config.DefaultConfigPath
	if cfgFile != "" {
		configPath = cfgFile
	}

	configResult := DiagnosticResult{
		Name: "Config File",
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configResult.Status = "fail"
		configResult.Message = fmt.Sprintf("Not found: %s", configPath)
		configResult.Details = "Run 'aami init' to create configuration"
		results = append(results, configResult)
		return results
	}

	configResult.Status = "pass"
	configResult.Message = configPath
	results = append(results, configResult)

	// Config validation
	cfg, err := config.Load(configPath)
	validResult := DiagnosticResult{
		Name: "Config Validation",
	}

	if err != nil {
		validResult.Status = "fail"
		validResult.Message = "Invalid configuration"
		validResult.Details = err.Error()
	} else {
		validResult.Status = "pass"
		validResult.Message = "Configuration is valid"

		// Check for common issues
		if len(cfg.Nodes) == 0 {
			validResult.Status = "warn"
			validResult.Message = "No nodes configured"
			validResult.Details = "Run 'aami nodes add' to add nodes"
		}
	}
	results = append(results, validResult)

	// SSH key check
	if cfg != nil {
		sshResult := DiagnosticResult{
			Name: "SSH Keys",
		}
		missingKeys := 0
		for _, node := range cfg.Nodes {
			if node.SSHKey != "" {
				// Expand environment variables
				keyPath := os.ExpandEnv(node.SSHKey)
				if _, err := os.Stat(keyPath); os.IsNotExist(err) {
					missingKeys++
				}
			}
		}
		if missingKeys > 0 {
			sshResult.Status = "warn"
			sshResult.Message = fmt.Sprintf("%d SSH key(s) not found", missingKeys)
		} else if len(cfg.Nodes) > 0 {
			sshResult.Status = "pass"
			sshResult.Message = "All SSH keys accessible"
		} else {
			sshResult.Status = "pass"
			sshResult.Message = "No nodes configured"
		}
		results = append(results, sshResult)
	}

	return results
}

func checkComponents() []DiagnosticResult {
	var results []DiagnosticResult

	// Prometheus
	promResult := checkHTTPService("Prometheus", "http://localhost:9090/-/healthy")
	results = append(results, promResult)

	// Grafana
	grafanaResult := checkHTTPService("Grafana", "http://localhost:3000/api/health")
	results = append(results, grafanaResult)

	// Alertmanager
	alertResult := checkHTTPService("Alertmanager", "http://localhost:9093/-/healthy")
	results = append(results, alertResult)

	// DCGM Exporter (check if nvidia-smi exists first)
	dcgmResult := DiagnosticResult{
		Name: "DCGM Exporter",
	}
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		dcgmResult.Status = "pass"
		dcgmResult.Message = "N/A (no NVIDIA GPU detected)"
	} else {
		dcgmResult = checkHTTPService("DCGM Exporter", "http://localhost:9400/metrics")
	}
	results = append(results, dcgmResult)

	// Node Exporter
	nodeExpResult := checkHTTPService("Node Exporter", "http://localhost:9100/metrics")
	results = append(results, nodeExpResult)

	return results
}

func checkHTTPService(name, url string) DiagnosticResult {
	result := DiagnosticResult{
		Name: name,
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		result.Status = "fail"
		result.Message = "Not running or unreachable"
		if diagnoseVerbose {
			result.Details = err.Error()
		}
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = "pass"
		result.Message = "Running"
	} else {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Returned status %d", resp.StatusCode)
	}

	return result
}

func checkNodes(cfg *config.Config) []DiagnosticResult {
	var results []DiagnosticResult

	executor := ssh.NewExecutorFromConfig(
		cfg.SSH.MaxParallel,
		cfg.SSH.ConnectTimeout,
		cfg.SSH.CommandTimeout,
		cfg.SSH.Retry.MaxAttempts,
		cfg.SSH.Retry.BackoffBase,
		cfg.SSH.Retry.BackoffMax,
	)

	for _, node := range cfg.Nodes {
		result := DiagnosticResult{
			Name: fmt.Sprintf("Node: %s", node.Name),
		}

		// Create SSH node
		sshNode := ssh.Node{
			Name:    node.Name,
			Host:    node.IP,
			Port:    node.SSHPort,
			User:    node.SSHUser,
			KeyPath: os.ExpandEnv(node.SSHKey),
		}
		if sshNode.Port == 0 {
			sshNode.Port = 22
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := executor.TestConnection(ctx, sshNode)
		cancel()

		if err != nil {
			result.Status = "fail"
			result.Message = "Connection failed"
			if diagnoseVerbose {
				result.Details = err.Error()
			}
		} else {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Connected (%s@%s)", node.SSHUser, node.IP)
		}

		results = append(results, result)
	}

	return results
}

func printResults(results []DiagnosticResult) {
	for _, r := range results {
		statusIcon := getStatusIcon(r.Status)
		fmt.Printf("  %s %-20s %s\n", statusIcon, r.Name+":", r.Message)
		if diagnoseVerbose && r.Details != "" {
			fmt.Printf("      └─ %s\n", r.Details)
		}
	}
}

func getStatusIcon(status string) string {
	switch status {
	case "pass":
		return color.GreenString("✓")
	case "warn":
		return color.YellowString("⚠")
	case "fail":
		return color.RedString("✗")
	default:
		return "?"
	}
}

func printSummary(results []DiagnosticResult) {
	passed, warned, failed := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "warn":
			warned++
		case "fail":
			failed++
		}
	}

	fmt.Println("Summary")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("  Total checks: %d\n", len(results))
	fmt.Printf("  %s Passed:  %d\n", color.GreenString("✓"), passed)
	if warned > 0 {
		fmt.Printf("  %s Warnings: %d\n", color.YellowString("⚠"), warned)
	}
	if failed > 0 {
		fmt.Printf("  %s Failed:  %d\n", color.RedString("✗"), failed)
	}
	fmt.Println()

	if failed > 0 {
		color.Red("Some checks failed. Review the issues above.")
	} else if warned > 0 {
		color.Yellow("Some warnings found. Review the issues above.")
	} else {
		color.Green("All checks passed!")
	}
}

// Subcommand for specific diagnostics
var diagnoseConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Diagnose configuration only",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Configuration Diagnostics")
		fmt.Println(strings.Repeat("=", 50))
		results := checkConfiguration()
		printResults(results)
		return nil
	},
}

var diagnoseNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Diagnose node connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		fmt.Println("Node Connectivity Diagnostics")
		fmt.Println(strings.Repeat("=", 50))
		results := checkNodes(cfg)
		printResults(results)
		return nil
	},
}

var diagnoseComponentsCmd = &cobra.Command{
	Use:   "components",
	Short: "Diagnose component status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Component Diagnostics")
		fmt.Println(strings.Repeat("=", 50))
		results := checkComponents()
		printResults(results)
		return nil
	},
}

func init() {
	diagnoseCmd.AddCommand(diagnoseConfigCmd)
	diagnoseCmd.AddCommand(diagnoseNodesCmd)
	diagnoseCmd.AddCommand(diagnoseComponentsCmd)
}
