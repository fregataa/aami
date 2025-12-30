package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/pkg/jobmanager"
)

// JobResponse represents a job in API responses
type JobResponse struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Result    any        `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
	Progress  int        `json:"progress,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Duration  string     `json:"duration,omitempty"`
}

// ToJobResponse converts a jobmanager.Job to JobResponse
func ToJobResponse(job *jobmanager.Job) JobResponse {
	resp := JobResponse{
		ID:        job.ID,
		Type:      job.Type,
		Status:    string(job.Status),
		Result:    job.Result,
		Error:     job.Error,
		Progress:  job.Progress,
		CreatedAt: job.CreatedAt,
		StartedAt: job.StartedAt,
		EndedAt:   job.EndedAt,
	}

	// Include human-readable duration if job has ended
	if job.Duration() > 0 {
		resp.Duration = job.Duration().String()
	}

	return resp
}

// ToJobResponseList converts a slice of jobs to JobResponse slice
func ToJobResponseList(jobs []*jobmanager.Job) []JobResponse {
	responses := make([]JobResponse, len(jobs))
	for i, job := range jobs {
		responses[i] = ToJobResponse(job)
	}
	return responses
}

// JobListResponse represents the response for listing jobs
type JobListResponse struct {
	Jobs  []JobResponse `json:"jobs"`
	Total int           `json:"total"`
}

// SubmitJobResponse is returned when a job is submitted (202 Accepted)
type SubmitJobResponse struct {
	JobID   string `json:"job_id"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	PollURL string `json:"poll_url"`
	IsNew   bool   `json:"is_new"` // false if returning existing job (SubmitUnique)
}

// JobStatsResponse represents job manager statistics
type JobStatsResponse struct {
	TotalJobs   int `json:"total_jobs"`
	QueueLength int `json:"queue_length"`
	MaxWorkers  int `json:"max_workers"`
}

// ListJobsRequest represents query parameters for listing jobs
type ListJobsRequest struct {
	Type        string   `form:"type"`
	Status      []string `form:"status"`
	Limit       int      `form:"limit,default=20"`
	Offset      int      `form:"offset,default=0"`
	OldestFirst bool     `form:"oldest_first"`
}

// ToListOptions converts ListJobsRequest to jobmanager.ListOptions
func (r *ListJobsRequest) ToListOptions() jobmanager.ListOptions {
	opts := jobmanager.ListOptions{
		Type:        r.Type,
		Limit:       r.Limit,
		Offset:      r.Offset,
		OldestFirst: r.OldestFirst,
	}

	// Convert status strings to JobStatus
	for _, s := range r.Status {
		opts.Status = append(opts.Status, jobmanager.JobStatus(s))
	}

	return opts
}
