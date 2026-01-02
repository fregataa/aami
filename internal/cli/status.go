package cli

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cluster status",
	Long: `Display the status of the AAMI monitoring stack.

Shows:
  - Cluster configuration
  - Node count
  - Component health (Prometheus, Alertmanager, Grafana)`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println()
	fmt.Printf("%s\n", bold("AAMI Status"))
	fmt.Println(strings.Repeat("─", 40))

	// Cluster info
	fmt.Printf("\n%s\n", bold("Cluster"))
	fmt.Printf("  Name:  %s\n", cfg.Cluster.Name)
	fmt.Printf("  Nodes: %d\n", len(cfg.Nodes))

	// Alert presets
	if len(cfg.Alerts.Presets) > 0 {
		fmt.Printf("  Alert Presets: %s\n", strings.Join(cfg.Alerts.Presets, ", "))
	}

	// Components
	fmt.Printf("\n%s\n", bold("Components"))

	promURL := fmt.Sprintf("http://localhost:%d/-/ready", cfg.Prometheus.Port)
	checkComponent("Prometheus", promURL, cfg.Prometheus.Port, green, red, yellow)

	checkComponent("Alertmanager", "http://localhost:9093/-/ready", 9093, green, red, yellow)

	grafanaURL := fmt.Sprintf("http://localhost:%d/api/health", cfg.Grafana.Port)
	checkComponent("Grafana", grafanaURL, cfg.Grafana.Port, green, red, yellow)

	// Notifications
	fmt.Printf("\n%s\n", bold("Notifications"))
	if cfg.Notifications.Slack != nil && cfg.Notifications.Slack.Enabled {
		fmt.Printf("  Slack:   %s (%s)\n", green("enabled"), cfg.Notifications.Slack.Channel)
	} else {
		fmt.Printf("  Slack:   %s\n", yellow("disabled"))
	}

	if cfg.Notifications.Email != nil && cfg.Notifications.Email.Enabled {
		fmt.Printf("  Email:   %s\n", green("enabled"))
	} else {
		fmt.Printf("  Email:   %s\n", yellow("disabled"))
	}

	if cfg.Notifications.Webhook != nil && cfg.Notifications.Webhook.Enabled {
		fmt.Printf("  Webhook: %s\n", green("enabled"))
	} else {
		fmt.Printf("  Webhook: %s\n", yellow("disabled"))
	}

	fmt.Println()

	return nil
}

func checkComponent(name, url string, port int, green, red, yellow func(a ...interface{}) string) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)

	status := ""
	if err != nil {
		status = red("not running")
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			status = green("running")
		} else {
			status = yellow(fmt.Sprintf("unhealthy (%d)", resp.StatusCode))
		}
	}

	fmt.Printf("  %-12s %s (port %d)\n", name+":", status, port)
}
