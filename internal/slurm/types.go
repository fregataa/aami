package slurm

import "time"

// JobState represents the state of a Slurm job.
type JobState string

const (
	JobStatePending    JobState = "PENDING"
	JobStateRunning    JobState = "RUNNING"
	JobStateCompleted  JobState = "COMPLETED"
	JobStateFailed     JobState = "FAILED"
	JobStateCancelled  JobState = "CANCELLED"
	JobStateTimeout    JobState = "TIMEOUT"
	JobStateNodeFail   JobState = "NODE_FAIL"
	JobStatePreempted  JobState = "PREEMPTED"
	JobStateSuspended  JobState = "SUSPENDED"
	JobStateOutOfMem   JobState = "OUT_OF_MEMORY"
)

// Job represents a Slurm job.
type Job struct {
	ID          int64            `json:"job_id"`
	Name        string           `json:"name"`
	User        string           `json:"user"`
	Group       string           `json:"group"`
	Partition   string           `json:"partition"`
	State       JobState         `json:"state"`
	ExitCode    int              `json:"exit_code"`
	Nodes       []string         `json:"nodes"`
	NodeCount   int              `json:"node_count"`
	GPUs        []GPUAllocation  `json:"gpus"`
	GPUCount    int              `json:"gpu_count"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	SubmitTime  time.Time        `json:"submit_time"`
	TimeLimit   time.Duration    `json:"time_limit"`
	WorkDir     string           `json:"work_dir"`
	Command     string           `json:"command"`
	StdOut      string           `json:"stdout"`
	StdErr      string           `json:"stderr"`
	Account     string           `json:"account"`
	QOS         string           `json:"qos"`
	Priority    int              `json:"priority"`
	Reason      string           `json:"reason"`       // Reason for pending/failed
	Features    []string         `json:"features"`     // Required node features
	Constraints string           `json:"constraints"`  // GRES constraints
}

// GPUAllocation represents a GPU allocated to a job.
type GPUAllocation struct {
	Node     string `json:"node"`
	GPUIndex int    `json:"gpu_index"`
	UUID     string `json:"uuid"`
	Type     string `json:"type"` // e.g., "nvidia_a100"
}

// GPUEvent represents a GPU-related event during job execution.
type GPUEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Node      string    `json:"node"`
	GPUIndex  int       `json:"gpu_index"`
	Type      string    `json:"type"`     // "xid", "temperature", "ecc", "power", "throttle"
	Value     string    `json:"value"`
	Severity  string    `json:"severity"` // "info", "warning", "critical"
	Message   string    `json:"message"`
}

// CorrelationType indicates the strength of job-GPU event correlation.
type CorrelationType string

const (
	CorrelationNone      CorrelationType = "none"
	CorrelationUnlikely  CorrelationType = "unlikely"
	CorrelationPossible  CorrelationType = "possible"
	CorrelationLikely    CorrelationType = "likely"
	CorrelationConfirmed CorrelationType = "confirmed"
)

// JobGPUCorrelation represents the analysis of a job's relation to GPU events.
type JobGPUCorrelation struct {
	Job            Job             `json:"job"`
	GPUEvents      []GPUEvent      `json:"gpu_events"`
	Correlation    CorrelationType `json:"correlation"`
	Confidence     float64         `json:"confidence"`      // 0.0 to 1.0
	Recommendation string          `json:"recommendation"`
	AffectedGPUs   []GPUAllocation `json:"affected_gpus"`
	Summary        string          `json:"summary"`
}

// SlurmConfig holds Slurm integration configuration.
type SlurmConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Endpoint      string `yaml:"endpoint"`        // slurmrestd endpoint (optional)
	AuthToken     string `yaml:"auth_token"`      // JWT or API token
	PreJobCheck   bool   `yaml:"pre_job_check"`   // Check GPU before job starts
	PostJobCheck  bool   `yaml:"post_job_check"`  // Check GPU after job ends
	AutoDrain     bool   `yaml:"auto_drain"`      // Drain node on GPU issue
	DrainOnXid    []int  `yaml:"drain_on_xid"`    // Xid codes that trigger drain
	HealthThreshold int  `yaml:"health_threshold"` // Minimum health score (0-100)
}

// DefaultSlurmConfig returns default Slurm configuration.
func DefaultSlurmConfig() SlurmConfig {
	return SlurmConfig{
		Enabled:       false,
		PreJobCheck:   true,
		PostJobCheck:  true,
		AutoDrain:     false,
		HealthThreshold: 50,
		DrainOnXid: []int{
			31, // GPU memory page fault
			43, // GPU stopped processing
			45, // Preemptive cleanup
			48, // Double bit ECC error
			61, // Internal micro-controller halt
			62, // Internal micro-controller halt
			63, // ECC page retirement/row remap failure
			64, // ECC page retirement/row remapping recording failure
			68, // NVDEC error
			69, // Graphics engine exception
			74, // NVLINK error
			79, // GPU fallen off the bus
			92, // High single bit ECC error rate
			94, // Contained ECC error
			95, // Uncontained ECC error
		},
	}
}

// NodeDrainRequest represents a request to drain a node.
type NodeDrainRequest struct {
	Node     string    `json:"node"`
	Reason   string    `json:"reason"`
	JobID    int64     `json:"job_id,omitempty"`
	GPUIndex int       `json:"gpu_index,omitempty"`
	XidCode  int       `json:"xid_code,omitempty"`
	DrainAt  time.Time `json:"drain_at"`
}

// NodeState represents the state of a Slurm node.
type NodeState string

const (
	NodeStateIdle      NodeState = "IDLE"
	NodeStateAllocated NodeState = "ALLOCATED"
	NodeStateMixed     NodeState = "MIXED"
	NodeStateDrain     NodeState = "DRAIN"
	NodeStateDraining  NodeState = "DRAINING"
	NodeStateDown      NodeState = "DOWN"
	NodeStateError     NodeState = "ERROR"
	NodeStateFuture    NodeState = "FUTURE"
	NodeStateUnknown   NodeState = "UNKNOWN"
)

// NodeInfo represents Slurm node information.
type NodeInfo struct {
	Name        string    `json:"name"`
	State       NodeState `json:"state"`
	CPUs        int       `json:"cpus"`
	CPUsAlloc   int       `json:"cpus_alloc"`
	Memory      int64     `json:"memory"`       // MB
	MemoryAlloc int64     `json:"memory_alloc"` // MB
	GPUs        int       `json:"gpus"`
	GPUsAlloc   int       `json:"gpus_alloc"`
	Partitions  []string  `json:"partitions"`
	Features    []string  `json:"features"`
	Reason      string    `json:"reason"` // Drain reason if applicable
	Weight      int       `json:"weight"`
}

// PartitionInfo represents Slurm partition information.
type PartitionInfo struct {
	Name       string   `json:"name"`
	State      string   `json:"state"` // UP, DOWN, DRAIN, INACTIVE
	TotalNodes int      `json:"total_nodes"`
	IdleNodes  int      `json:"idle_nodes"`
	AllocNodes int      `json:"alloc_nodes"`
	DownNodes  int      `json:"down_nodes"`
	TotalCPUs  int      `json:"total_cpus"`
	TotalGPUs  int      `json:"total_gpus"`
	MaxTime    string   `json:"max_time"`
	DefaultTime string  `json:"default_time"`
	Nodes      []string `json:"nodes"`
}

// JobFilter defines filters for job queries.
type JobFilter struct {
	User      string     `json:"user,omitempty"`
	Partition string     `json:"partition,omitempty"`
	State     JobState   `json:"state,omitempty"`
	StartTime time.Time  `json:"start_time,omitempty"`
	EndTime   time.Time  `json:"end_time,omitempty"`
	Node      string     `json:"node,omitempty"`
	Account   string     `json:"account,omitempty"`
}

// JobStatistics represents job statistics for a time period.
type JobStatistics struct {
	Period        string  `json:"period"`
	TotalJobs     int     `json:"total_jobs"`
	CompletedJobs int     `json:"completed_jobs"`
	FailedJobs    int     `json:"failed_jobs"`
	GPUHours      float64 `json:"gpu_hours"`
	AvgWaitTime   float64 `json:"avg_wait_time_seconds"`
	AvgRunTime    float64 `json:"avg_run_time_seconds"`
	Efficiency    float64 `json:"efficiency"` // GPU utilization during jobs
}

// CorrelationLog represents a logged correlation event.
type CorrelationLog struct {
	Timestamp    time.Time       `json:"timestamp"`
	JobID        int64           `json:"job_id"`
	Node         string          `json:"node"`
	GPUIndex     int             `json:"gpu_index"`
	HealthScore  int             `json:"health_score"`
	ExitCode     int             `json:"exit_code"`
	Correlation  CorrelationType `json:"correlation"`
	EventType    string          `json:"event_type"`
	EventValue   string          `json:"event_value"`
	ActionTaken  string          `json:"action_taken"`
}
