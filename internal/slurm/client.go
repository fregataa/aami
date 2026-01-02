package slurm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Client provides access to Slurm functionality.
type Client struct {
	config     SlurmConfig
	httpClient *http.Client
}

// NewClient creates a new Slurm client.
func NewClient(cfg SlurmConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetJob retrieves job information by ID.
func (c *Client) GetJob(ctx context.Context, jobID int64) (*Job, error) {
	if c.config.Endpoint != "" {
		return c.getJobREST(ctx, jobID)
	}
	return c.getJobCLI(ctx, jobID)
}

// getJobREST retrieves job via slurmrestd API.
func (c *Client) getJobREST(ctx context.Context, jobID int64) (*Job, error) {
	url := fmt.Sprintf("%s/slurm/v0.0.40/job/%d", c.config.Endpoint, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.config.AuthToken != "" {
		req.Header.Set("X-SLURM-USER-TOKEN", c.config.AuthToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slurm API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("slurm API returned status %d", resp.StatusCode)
	}

	var result struct {
		Jobs []struct {
			JobID       int64  `json:"job_id"`
			Name        string `json:"name"`
			UserName    string `json:"user_name"`
			GroupName   string `json:"group_name"`
			Partition   string `json:"partition"`
			JobState    string `json:"job_state"`
			ExitCode    int    `json:"exit_code"`
			Nodes       string `json:"nodes"`
			StartTime   int64  `json:"start_time"`
			EndTime     int64  `json:"end_time"`
			SubmitTime  int64  `json:"submit_time"`
			TimeLimit   int    `json:"time_limit"`
			WorkDir     string `json:"work_dir"`
			Command     string `json:"command"`
			Account     string `json:"account"`
			QOS         string `json:"qos"`
			Priority    int    `json:"priority"`
			StateReason string `json:"state_reason"`
		} `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Jobs) == 0 {
		return nil, fmt.Errorf("job not found: %d", jobID)
	}

	j := result.Jobs[0]
	return &Job{
		ID:         j.JobID,
		Name:       j.Name,
		User:       j.UserName,
		Group:      j.GroupName,
		Partition:  j.Partition,
		State:      JobState(j.JobState),
		ExitCode:   j.ExitCode,
		Nodes:      c.expandNodeList(ctx, j.Nodes),
		StartTime:  time.Unix(j.StartTime, 0),
		EndTime:    time.Unix(j.EndTime, 0),
		SubmitTime: time.Unix(j.SubmitTime, 0),
		TimeLimit:  time.Duration(j.TimeLimit) * time.Minute,
		WorkDir:    j.WorkDir,
		Command:    j.Command,
		Account:    j.Account,
		QOS:        j.QOS,
		Priority:   j.Priority,
		Reason:     j.StateReason,
	}, nil
}

// getJobCLI retrieves job via scontrol command.
func (c *Client) getJobCLI(ctx context.Context, jobID int64) (*Job, error) {
	cmd := exec.CommandContext(ctx, "scontrol", "show", "job", strconv.FormatInt(jobID, 10))
	output, err := cmd.Output()
	if err != nil {
		// Try sacct for completed jobs
		return c.getJobFromSacct(ctx, jobID)
	}

	return c.parseScontrolJobOutput(ctx, string(output))
}

// getJobFromSacct retrieves completed job info from sacct.
func (c *Client) getJobFromSacct(ctx context.Context, jobID int64) (*Job, error) {
	cmd := exec.CommandContext(ctx, "sacct",
		"-j", strconv.FormatInt(jobID, 10),
		"--format=JobID,JobName,User,Group,Partition,State,ExitCode,NodeList,Start,End,Submit,Account,QOS",
		"--noheader", "--parsable2")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sacct failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("job not found: %d", jobID)
	}

	// First line is the main job entry
	parts := strings.Split(lines[0], "|")
	if len(parts) < 13 {
		return nil, fmt.Errorf("unexpected sacct output format")
	}

	job := &Job{
		User:      parts[2],
		Group:     parts[3],
		Partition: parts[4],
		State:     JobState(parts[5]),
		Nodes:     c.expandNodeList(ctx, parts[7]),
		Account:   parts[11],
		QOS:       parts[12],
	}

	// Parse job ID (may have .batch suffix)
	idStr := strings.Split(parts[0], ".")[0]
	job.ID, _ = strconv.ParseInt(idStr, 10, 64)
	job.Name = parts[1]

	// Parse exit code (format: exitcode:signal)
	exitParts := strings.Split(parts[6], ":")
	if len(exitParts) > 0 {
		job.ExitCode, _ = strconv.Atoi(exitParts[0])
	}

	// Parse times
	job.StartTime = parseSlurTime(parts[8])
	job.EndTime = parseSlurTime(parts[9])
	job.SubmitTime = parseSlurTime(parts[10])

	return job, nil
}

// parseScontrolJobOutput parses scontrol show job output.
func (c *Client) parseScontrolJobOutput(ctx context.Context, output string) (*Job, error) {
	job := &Job{}

	// scontrol output is key=value pairs
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		for _, part := range strings.Fields(line) {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key, value := kv[0], kv[1]

			switch key {
			case "JobId":
				job.ID, _ = strconv.ParseInt(value, 10, 64)
			case "JobName":
				job.Name = value
			case "UserId":
				// Format: user(uid)
				job.User = strings.Split(value, "(")[0]
			case "GroupId":
				job.Group = strings.Split(value, "(")[0]
			case "Partition":
				job.Partition = value
			case "JobState":
				job.State = JobState(value)
			case "ExitCode":
				parts := strings.Split(value, ":")
				if len(parts) > 0 {
					job.ExitCode, _ = strconv.Atoi(parts[0])
				}
			case "NodeList":
				job.Nodes = c.expandNodeList(ctx, value)
			case "NumNodes":
				job.NodeCount, _ = strconv.Atoi(value)
			case "StartTime":
				job.StartTime = parseSlurTime(value)
			case "EndTime":
				job.EndTime = parseSlurTime(value)
			case "SubmitTime":
				job.SubmitTime = parseSlurTime(value)
			case "TimeLimit":
				job.TimeLimit = parseTimeLimit(value)
			case "WorkDir":
				job.WorkDir = value
			case "Command":
				job.Command = value
			case "StdOut":
				job.StdOut = value
			case "StdErr":
				job.StdErr = value
			case "Account":
				job.Account = value
			case "QOS":
				job.QOS = value
			case "Priority":
				job.Priority, _ = strconv.Atoi(value)
			case "Reason":
				job.Reason = value
			case "Features":
				if value != "(null)" {
					job.Features = strings.Split(value, ",")
				}
			case "Gres":
				job.Constraints = value
				job.GPUCount = parseGPUCount(value)
			}
		}
	}

	if job.ID == 0 {
		return nil, fmt.Errorf("failed to parse job output")
	}

	return job, nil
}

// expandNodeList expands Slurm node list notation.
// e.g., "gpu-node-[01-04]" -> ["gpu-node-01", "gpu-node-02", "gpu-node-03", "gpu-node-04"]
func (c *Client) expandNodeList(ctx context.Context, nodeList string) []string {
	if nodeList == "" || nodeList == "(null)" {
		return nil
	}

	// Use scontrol to expand
	cmd := exec.CommandContext(ctx, "scontrol", "show", "hostnames", nodeList)
	output, err := cmd.Output()
	if err != nil {
		// Fallback: return as-is
		return []string{nodeList}
	}

	var nodes []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			nodes = append(nodes, line)
		}
	}
	return nodes
}

// GetRunningJobs returns all currently running jobs.
func (c *Client) GetRunningJobs(ctx context.Context) ([]Job, error) {
	return c.GetJobs(ctx, JobFilter{State: JobStateRunning})
}

// GetJobs retrieves jobs matching the filter.
func (c *Client) GetJobs(ctx context.Context, filter JobFilter) ([]Job, error) {
	args := []string{"-h", "-o", "%i|%j|%u|%P|%T|%N|%l|%V|%S|%e|%a"}

	if filter.User != "" {
		args = append(args, "-u", filter.User)
	}
	if filter.Partition != "" {
		args = append(args, "-p", filter.Partition)
	}
	if filter.State != "" {
		args = append(args, "-t", string(filter.State))
	}
	if filter.Node != "" {
		args = append(args, "-w", filter.Node)
	}
	if filter.Account != "" {
		args = append(args, "-A", filter.Account)
	}

	cmd := exec.CommandContext(ctx, "squeue", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("squeue failed: %w", err)
	}

	var jobs []Job
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 11 {
			continue
		}

		id, _ := strconv.ParseInt(parts[0], 10, 64)
		jobs = append(jobs, Job{
			ID:         id,
			Name:       parts[1],
			User:       parts[2],
			Partition:  parts[3],
			State:      JobState(parts[4]),
			Nodes:      c.expandNodeList(ctx, parts[5]),
			TimeLimit:  parseTimeLimit(parts[6]),
			SubmitTime: parseSlurTime(parts[7]),
			StartTime:  parseSlurTime(parts[8]),
			EndTime:    parseSlurTime(parts[9]),
			Account:    parts[10],
		})
	}

	return jobs, nil
}

// GetNode retrieves node information.
func (c *Client) GetNode(ctx context.Context, nodeName string) (*NodeInfo, error) {
	cmd := exec.CommandContext(ctx, "scontrol", "show", "node", nodeName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("scontrol show node failed: %w", err)
	}

	return c.parseNodeOutput(string(output))
}

func (c *Client) parseNodeOutput(output string) (*NodeInfo, error) {
	node := &NodeInfo{}

	for _, part := range strings.Fields(output) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]

		switch key {
		case "NodeName":
			node.Name = value
		case "State":
			node.State = NodeState(strings.Split(value, "+")[0])
		case "CPUTot":
			node.CPUs, _ = strconv.Atoi(value)
		case "CPUAlloc":
			node.CPUsAlloc, _ = strconv.Atoi(value)
		case "RealMemory":
			node.Memory, _ = strconv.ParseInt(value, 10, 64)
		case "AllocMem":
			node.MemoryAlloc, _ = strconv.ParseInt(value, 10, 64)
		case "Gres":
			node.GPUs = parseGPUCount(value)
		case "GresUsed":
			node.GPUsAlloc = parseGPUCount(value)
		case "Partitions":
			node.Partitions = strings.Split(value, ",")
		case "AvailableFeatures":
			if value != "(null)" {
				node.Features = strings.Split(value, ",")
			}
		case "Reason":
			if value != "(null)" {
				node.Reason = value
			}
		case "Weight":
			node.Weight, _ = strconv.Atoi(value)
		}
	}

	if node.Name == "" {
		return nil, fmt.Errorf("failed to parse node output")
	}

	return node, nil
}

// DrainNode drains a node from Slurm scheduling.
func (c *Client) DrainNode(ctx context.Context, nodeName, reason string) error {
	cmd := exec.CommandContext(ctx, "scontrol", "update",
		fmt.Sprintf("NodeName=%s", nodeName),
		"State=DRAIN",
		fmt.Sprintf("Reason=%s", reason))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("drain failed: %s", string(output))
	}

	return nil
}

// ResumeNode resumes a drained node.
func (c *Client) ResumeNode(ctx context.Context, nodeName string) error {
	cmd := exec.CommandContext(ctx, "scontrol", "update",
		fmt.Sprintf("NodeName=%s", nodeName),
		"State=RESUME")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resume failed: %s", string(output))
	}

	return nil
}

// GetPartitions retrieves all partition information.
func (c *Client) GetPartitions(ctx context.Context) ([]PartitionInfo, error) {
	cmd := exec.CommandContext(ctx, "sinfo", "-h", "-o", "%P|%a|%D|%T|%C|%G")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sinfo failed: %w", err)
	}

	partitionMap := make(map[string]*PartitionInfo)

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		name := strings.TrimSuffix(parts[0], "*") // Remove default marker
		partition, exists := partitionMap[name]
		if !exists {
			partition = &PartitionInfo{
				Name:  name,
				State: parts[1],
			}
			partitionMap[name] = partition
		}

		nodes, _ := strconv.Atoi(parts[2])
		partition.TotalNodes += nodes

		// Parse node state
		state := parts[3]
		if strings.Contains(state, "idle") {
			partition.IdleNodes += nodes
		} else if strings.Contains(state, "alloc") || strings.Contains(state, "mix") {
			partition.AllocNodes += nodes
		} else if strings.Contains(state, "down") || strings.Contains(state, "drain") {
			partition.DownNodes += nodes
		}

		// Parse CPU info (A/I/O/T format)
		cpuParts := strings.Split(parts[4], "/")
		if len(cpuParts) >= 4 {
			total, _ := strconv.Atoi(cpuParts[3])
			partition.TotalCPUs += total
		}

		// Parse GRES (GPU count)
		partition.TotalGPUs += parseGPUCount(parts[5])
	}

	var partitions []PartitionInfo
	for _, p := range partitionMap {
		partitions = append(partitions, *p)
	}

	return partitions, nil
}

// GetJobsByNode returns jobs running on a specific node.
func (c *Client) GetJobsByNode(ctx context.Context, nodeName string) ([]Job, error) {
	return c.GetJobs(ctx, JobFilter{Node: nodeName})
}

// Helper functions

func parseSlurTime(s string) time.Time {
	if s == "" || s == "Unknown" || s == "N/A" {
		return time.Time{}
	}

	// Try various formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	return time.Time{}
}

func parseTimeLimit(s string) time.Duration {
	if s == "" || s == "UNLIMITED" {
		return 0
	}

	// Format: D-HH:MM:SS or HH:MM:SS or MM:SS
	parts := strings.Split(s, "-")
	var days int
	var timePart string

	if len(parts) == 2 {
		days, _ = strconv.Atoi(parts[0])
		timePart = parts[1]
	} else {
		timePart = parts[0]
	}

	timeParts := strings.Split(timePart, ":")
	var hours, minutes, seconds int

	switch len(timeParts) {
	case 3:
		hours, _ = strconv.Atoi(timeParts[0])
		minutes, _ = strconv.Atoi(timeParts[1])
		seconds, _ = strconv.Atoi(timeParts[2])
	case 2:
		minutes, _ = strconv.Atoi(timeParts[0])
		seconds, _ = strconv.Atoi(timeParts[1])
	case 1:
		minutes, _ = strconv.Atoi(timeParts[0])
	}

	return time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second
}

func parseGPUCount(gres string) int {
	if gres == "" || gres == "(null)" {
		return 0
	}

	// Format: gpu:type:count or gpu:count
	for _, part := range strings.Split(gres, ",") {
		if strings.HasPrefix(part, "gpu") {
			fields := strings.Split(part, ":")
			if len(fields) >= 2 {
				// Last field is the count
				count, _ := strconv.Atoi(fields[len(fields)-1])
				return count
			}
		}
	}

	return 0
}
