package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/slurm"
)

var (
	slurmDrainReason   string
	slurmOutputJSON    bool
	slurmAnalyzeHours  int
	slurmInstallForce  bool
)

var slurmCmd = &cobra.Command{
	Use:   "slurm",
	Short: "Slurm integration for GPU-job correlation",
	Long: `Integrate AAMI with Slurm for GPU health monitoring and job correlation.

Features:
  - Analyze failed jobs for GPU-related issues
  - Drain nodes with GPU problems
  - Install prolog/epilog hooks for automatic checks
  - View job-GPU correlation history

Examples:
  aami slurm job-analyze 12345        # Analyze job for GPU issues
  aami slurm drain gpu-node-01        # Drain a node
  aami slurm install-hooks            # Install Slurm hooks
  aami slurm jobs --node gpu-node-01  # List jobs on a node`,
}

var slurmJobAnalyzeCmd = &cobra.Command{
	Use:   "job-analyze <job-id>",
	Short: "Analyze job for GPU-related issues",
	Long: `Analyze a Slurm job and correlate with GPU events.

This queries Prometheus for GPU events (Xid errors, temperature spikes,
ECC errors) that occurred during the job's execution and determines
if the job failure may be GPU-related.

Examples:
  aami slurm job-analyze 12345
  aami slurm job-analyze 12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSlurmJobAnalyze,
}

var slurmDrainCmd = &cobra.Command{
	Use:   "drain <node>",
	Short: "Drain a node from Slurm scheduling",
	Long: `Mark a node as DRAIN in Slurm, preventing new jobs from starting.

Existing jobs will continue running until completion. Use this when
a GPU issue is detected that requires investigation.

Examples:
  aami slurm drain gpu-node-01
  aami slurm drain gpu-node-01 --reason "GPU maintenance"`,
	Args: cobra.ExactArgs(1),
	RunE: runSlurmDrain,
}

var slurmResumeCmd = &cobra.Command{
	Use:   "resume <node>",
	Short: "Resume a drained node",
	Long:  `Remove the DRAIN state from a node, allowing new jobs to be scheduled.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSlurmResume,
}

var slurmInstallHooksCmd = &cobra.Command{
	Use:   "install-hooks",
	Short: "Install Slurm prolog/epilog hooks",
	Long: `Install AAMI health check hooks for Slurm.

This creates prolog and epilog scripts that:
  - Check GPU health before jobs start (prolog)
  - Log correlation data after jobs complete (epilog)
  - Optionally auto-drain nodes with GPU issues

The hooks are installed to /etc/slurm/ by default.`,
	RunE: runSlurmInstallHooks,
}

var slurmUninstallHooksCmd = &cobra.Command{
	Use:   "uninstall-hooks",
	Short: "Remove installed Slurm hooks",
	RunE:  runSlurmUninstallHooks,
}

var slurmJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "List Slurm jobs",
	Long: `List running or recent Slurm jobs with optional filtering.

Examples:
  aami slurm jobs                      # All running jobs
  aami slurm jobs --node gpu-node-01   # Jobs on specific node
  aami slurm jobs --user alice         # Jobs by user`,
	RunE: runSlurmJobs,
}

var slurmNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List Slurm node status",
	RunE:  runSlurmNodes,
}

var slurmLogCorrelationCmd = &cobra.Command{
	Use:    "log-correlation",
	Short:  "Log job-GPU correlation event (internal use)",
	Hidden: true, // Called by hooks
	RunE:   runSlurmLogCorrelation,
}

var slurmNodeAnalyzeCmd = &cobra.Command{
	Use:   "node-analyze <node>",
	Short: "Analyze recent jobs on a node for GPU issues",
	Args:  cobra.ExactArgs(1),
	RunE:  runSlurmNodeAnalyze,
}

var (
	slurmJobsNode      string
	slurmJobsUser      string
	slurmJobsPartition string
	slurmLogJobID      int64
	slurmLogNode       string
	slurmLogScore      int
	slurmLogExitCode   int
)

func init() {
	rootCmd.AddCommand(slurmCmd)

	// job-analyze
	slurmJobAnalyzeCmd.Flags().BoolVar(&slurmOutputJSON, "json", false, "Output in JSON format")
	slurmCmd.AddCommand(slurmJobAnalyzeCmd)

	// drain
	slurmDrainCmd.Flags().StringVar(&slurmDrainReason, "reason", "AAMI: GPU health issue",
		"Reason for draining the node")
	slurmCmd.AddCommand(slurmDrainCmd)

	// resume
	slurmCmd.AddCommand(slurmResumeCmd)

	// install-hooks
	slurmInstallHooksCmd.Flags().BoolVar(&slurmInstallForce, "force", false,
		"Overwrite existing hooks")
	slurmCmd.AddCommand(slurmInstallHooksCmd)

	// uninstall-hooks
	slurmCmd.AddCommand(slurmUninstallHooksCmd)

	// jobs
	slurmJobsCmd.Flags().StringVar(&slurmJobsNode, "node", "", "Filter by node")
	slurmJobsCmd.Flags().StringVar(&slurmJobsUser, "user", "", "Filter by user")
	slurmJobsCmd.Flags().StringVar(&slurmJobsPartition, "partition", "", "Filter by partition")
	slurmCmd.AddCommand(slurmJobsCmd)

	// nodes
	slurmCmd.AddCommand(slurmNodesCmd)

	// log-correlation (hidden, for hooks)
	slurmLogCorrelationCmd.Flags().Int64Var(&slurmLogJobID, "job", 0, "Job ID")
	slurmLogCorrelationCmd.Flags().StringVar(&slurmLogNode, "node", "", "Node name")
	slurmLogCorrelationCmd.Flags().IntVar(&slurmLogScore, "score", 0, "Health score")
	slurmLogCorrelationCmd.Flags().IntVar(&slurmLogExitCode, "exit-code", 0, "Exit code")
	slurmCmd.AddCommand(slurmLogCorrelationCmd)

	// node-analyze
	slurmNodeAnalyzeCmd.Flags().IntVar(&slurmAnalyzeHours, "hours", 24,
		"Hours of history to analyze")
	slurmCmd.AddCommand(slurmNodeAnalyzeCmd)
}

func runSlurmJobAnalyze(cmd *cobra.Command, args []string) error {
	jobID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid job ID: %s", args[0])
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	slurmClient := slurm.NewClient(slurm.DefaultSlurmConfig())
	prometheusURL := fmt.Sprintf("http://localhost:%d", cfg.Prometheus.Port)
	analyzer := slurm.NewAnalyzer(slurmClient, prometheusURL)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("Analyzing job %d...\n\n", jobID)

	result, err := analyzer.AnalyzeJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if slurmOutputJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Print results
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║              Job-GPU Correlation Analysis          ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	// Job info
	fmt.Println("Job Information")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("  Job ID:     %d\n", result.Job.ID)
	fmt.Printf("  Name:       %s\n", result.Job.Name)
	fmt.Printf("  User:       %s\n", result.Job.User)
	fmt.Printf("  Partition:  %s\n", result.Job.Partition)
	fmt.Printf("  State:      %s\n", colorJobState(result.Job.State))
	fmt.Printf("  Exit Code:  %d\n", result.Job.ExitCode)
	fmt.Printf("  Nodes:      %s\n", strings.Join(result.Job.Nodes, ", "))
	if !result.Job.StartTime.IsZero() {
		fmt.Printf("  Start:      %s\n", result.Job.StartTime.Format("2006-01-02 15:04:05"))
	}
	if !result.Job.EndTime.IsZero() {
		fmt.Printf("  End:        %s\n", result.Job.EndTime.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// GPU events
	fmt.Println("GPU Events During Job")
	fmt.Println(strings.Repeat("-", 50))

	if len(result.GPUEvents) == 0 {
		fmt.Println("  No GPU events detected")
	} else {
		for _, event := range result.GPUEvents {
			icon := "•"
			if event.Severity == "critical" {
				icon = color.RedString("✗")
			} else if event.Severity == "warning" {
				icon = color.YellowString("⚠")
			}
			fmt.Printf("  %s [%s] %s\n", icon,
				event.Timestamp.Format("15:04:05"),
				event.Message)
		}
	}
	fmt.Println()

	// Correlation result
	fmt.Println("Correlation Analysis")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("  Correlation:  %s\n", colorCorrelation(result.Correlation))
	fmt.Printf("  Confidence:   %.0f%%\n", result.Confidence*100)
	fmt.Printf("  Summary:      %s\n", result.Summary)
	fmt.Println()

	// Affected GPUs
	if len(result.AffectedGPUs) > 0 {
		fmt.Println("Affected GPUs")
		fmt.Println(strings.Repeat("-", 50))
		for _, gpu := range result.AffectedGPUs {
			fmt.Printf("  %s %s GPU %d\n", color.RedString("•"), gpu.Node, gpu.GPUIndex)
		}
		fmt.Println()
	}

	// Recommendations
	if result.Recommendation != "" && result.Recommendation != "No specific action required" {
		fmt.Println("Recommendations")
		fmt.Println(strings.Repeat("-", 50))
		for _, line := range strings.Split(result.Recommendation, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	return nil
}

func runSlurmDrain(cmd *cobra.Command, args []string) error {
	node := args[0]

	slurmClient := slurm.NewClient(slurm.DefaultSlurmConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Draining node %s...\n", node)

	if err := slurmClient.DrainNode(ctx, node, slurmDrainReason); err != nil {
		return fmt.Errorf("drain failed: %w", err)
	}

	color.Green("✓ Node %s drained", node)
	fmt.Printf("  Reason: %s\n", slurmDrainReason)
	fmt.Println()
	fmt.Println("To resume the node:")
	fmt.Printf("  aami slurm resume %s\n", node)

	return nil
}

func runSlurmResume(cmd *cobra.Command, args []string) error {
	node := args[0]

	slurmClient := slurm.NewClient(slurm.DefaultSlurmConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Resuming node %s...\n", node)

	if err := slurmClient.ResumeNode(ctx, node); err != nil {
		return fmt.Errorf("resume failed: %w", err)
	}

	color.Green("✓ Node %s resumed", node)

	return nil
}

func runSlurmInstallHooks(cmd *cobra.Command, args []string) error {
	prologPath := "/etc/slurm/aami-prolog.sh"
	epilogPath := "/etc/slurm/aami-epilog.sh"

	// Check if files exist
	if !slurmInstallForce {
		if _, err := os.Stat(prologPath); err == nil {
			return fmt.Errorf("prolog already exists: %s (use --force to overwrite)", prologPath)
		}
		if _, err := os.Stat(epilogPath); err == nil {
			return fmt.Errorf("epilog already exists: %s (use --force to overwrite)", epilogPath)
		}
	}

	slurmCfg := slurm.DefaultSlurmConfig()
	slurmCfg.PreJobCheck = true
	slurmCfg.PostJobCheck = true

	slurmClient := slurm.NewClient(slurmCfg)
	hookMgr := slurm.NewHookManager(slurmCfg, slurmClient)

	fmt.Println("Installing Slurm hooks...")

	if err := hookMgr.InstallHooks(prologPath, epilogPath); err != nil {
		return fmt.Errorf("install hooks: %w", err)
	}

	color.Green("✓ Hooks installed successfully")
	fmt.Println()
	fmt.Println("Installed files:")
	fmt.Printf("  Prolog: %s\n", prologPath)
	fmt.Printf("  Epilog: %s\n", epilogPath)
	fmt.Println()
	fmt.Println("Add the following to slurm.conf:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(hookMgr.GenerateSlurmConf(prologPath, epilogPath))
	fmt.Println()
	fmt.Println("Then restart slurmctld:")
	fmt.Println("  sudo systemctl restart slurmctld")

	return nil
}

func runSlurmUninstallHooks(cmd *cobra.Command, args []string) error {
	prologPath := "/etc/slurm/aami-prolog.sh"
	epilogPath := "/etc/slurm/aami-epilog.sh"

	hookMgr := slurm.NewHookManager(slurm.DefaultSlurmConfig(), nil)

	fmt.Println("Removing Slurm hooks...")

	if err := hookMgr.UninstallHooks(prologPath, epilogPath); err != nil {
		return fmt.Errorf("uninstall hooks: %w", err)
	}

	color.Green("✓ Hooks removed")
	fmt.Println()
	fmt.Println("Remember to remove the Prolog/Epilog lines from slurm.conf")
	fmt.Println("and restart slurmctld")

	return nil
}

func runSlurmJobs(cmd *cobra.Command, args []string) error {
	slurmClient := slurm.NewClient(slurm.DefaultSlurmConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := slurm.JobFilter{
		Node:      slurmJobsNode,
		User:      slurmJobsUser,
		Partition: slurmJobsPartition,
	}

	jobs, err := slurmClient.GetJobs(ctx, filter)
	if err != nil {
		return fmt.Errorf("get jobs: %w", err)
	}

	if len(jobs) == 0 {
		fmt.Println("No jobs found")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Job ID", "Name", "User", "Partition", "State", "Nodes", "Time"})
	table.SetBorder(false)

	for _, job := range jobs {
		runtime := ""
		if !job.StartTime.IsZero() {
			duration := time.Since(job.StartTime)
			if !job.EndTime.IsZero() {
				duration = job.EndTime.Sub(job.StartTime)
			}
			runtime = formatDuration(duration)
		}

		nodes := strings.Join(job.Nodes, ",")
		if len(nodes) > 20 {
			nodes = nodes[:17] + "..."
		}

		table.Append([]string{
			strconv.FormatInt(job.ID, 10),
			truncate(job.Name, 20),
			job.User,
			job.Partition,
			string(job.State),
			nodes,
			runtime,
		})
	}

	table.Render()

	return nil
}

func runSlurmNodes(cmd *cobra.Command, args []string) error {
	slurmClient := slurm.NewClient(slurm.DefaultSlurmConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	partitions, err := slurmClient.GetPartitions(ctx)
	if err != nil {
		return fmt.Errorf("get partitions: %w", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Partition", "State", "Nodes", "Idle", "Alloc", "Down", "GPUs"})
	table.SetBorder(false)

	for _, p := range partitions {
		table.Append([]string{
			p.Name,
			p.State,
			strconv.Itoa(p.TotalNodes),
			strconv.Itoa(p.IdleNodes),
			strconv.Itoa(p.AllocNodes),
			strconv.Itoa(p.DownNodes),
			strconv.Itoa(p.TotalGPUs),
		})
	}

	table.Render()

	return nil
}

func runSlurmLogCorrelation(cmd *cobra.Command, args []string) error {
	// This is called by the epilog hook to log correlation data
	log := slurm.CorrelationLog{
		Timestamp:   time.Now(),
		JobID:       slurmLogJobID,
		Node:        slurmLogNode,
		HealthScore: slurmLogScore,
		ExitCode:    slurmLogExitCode,
	}

	// In production, this would write to a database or log file
	data, _ := json.Marshal(log)
	fmt.Println(string(data))

	return nil
}

func runSlurmNodeAnalyze(cmd *cobra.Command, args []string) error {
	node := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	slurmClient := slurm.NewClient(slurm.DefaultSlurmConfig())
	prometheusURL := fmt.Sprintf("http://localhost:%d", cfg.Prometheus.Port)
	analyzer := slurm.NewAnalyzer(slurmClient, prometheusURL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Printf("Analyzing jobs on %s (last %d hours)...\n\n", node, slurmAnalyzeHours)

	correlations, err := analyzer.AnalyzeNode(ctx, node, slurmAnalyzeHours)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if len(correlations) == 0 {
		fmt.Println("No jobs found on this node in the specified time period")
		return nil
	}

	// Summary
	var withIssues, critical int
	for _, c := range correlations {
		if c.Correlation != slurm.CorrelationNone {
			withIssues++
		}
		if c.Correlation == slurm.CorrelationConfirmed {
			critical++
		}
	}

	fmt.Printf("Analyzed %d jobs: %d with GPU events, %d critical\n\n", len(correlations), withIssues, critical)

	// Table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Job ID", "User", "State", "Events", "Correlation", "Confidence"})
	table.SetBorder(false)

	for _, c := range correlations {
		if c.Correlation == slurm.CorrelationNone && len(correlations) > 10 {
			continue // Skip jobs without issues if there are many
		}

		table.Append([]string{
			strconv.FormatInt(c.Job.ID, 10),
			c.Job.User,
			string(c.Job.State),
			strconv.Itoa(len(c.GPUEvents)),
			string(c.Correlation),
			fmt.Sprintf("%.0f%%", c.Confidence*100),
		})
	}

	table.Render()

	return nil
}

// Helper functions

func colorJobState(state slurm.JobState) string {
	switch state {
	case slurm.JobStateCompleted:
		return color.GreenString(string(state))
	case slurm.JobStateFailed, slurm.JobStateNodeFail, slurm.JobStateTimeout:
		return color.RedString(string(state))
	case slurm.JobStateRunning:
		return color.CyanString(string(state))
	case slurm.JobStatePending:
		return color.YellowString(string(state))
	default:
		return string(state)
	}
}

func colorCorrelation(c slurm.CorrelationType) string {
	switch c {
	case slurm.CorrelationConfirmed:
		return color.RedString(string(c))
	case slurm.CorrelationLikely:
		return color.YellowString(string(c))
	case slurm.CorrelationPossible:
		return color.YellowString(string(c))
	case slurm.CorrelationNone:
		return color.GreenString(string(c))
	default:
		return string(c)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours < 24 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
