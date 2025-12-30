package handler

import (
	"errors"
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/pkg/jobmanager"
	"github.com/gin-gonic/gin"
)

// JobHandler handles HTTP requests for job management
type JobHandler struct {
	manager *jobmanager.Manager
}

// NewJobHandler creates a new JobHandler
func NewJobHandler(manager *jobmanager.Manager) *JobHandler {
	return &JobHandler{
		manager: manager,
	}
}

// GetByID handles GET /api/v1/jobs/:id
func (h *JobHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job id is required",
		})
		return
	}

	job, err := h.manager.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, jobmanager.ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "job not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToJobResponse(job))
}

// List handles GET /api/v1/jobs
func (h *JobHandler) List(c *gin.Context) {
	var req dto.ListJobsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	jobs, err := h.manager.List(c.Request.Context(), req.ToListOptions())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.JobListResponse{
		Jobs:  dto.ToJobResponseList(jobs),
		Total: len(jobs),
	})
}

// Cancel handles DELETE /api/v1/jobs/:id
func (h *JobHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job id is required",
		})
		return
	}

	err := h.manager.Cancel(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, jobmanager.ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "job not found",
			})
			return
		}
		if errors.Is(err, jobmanager.ErrJobNotCancellable) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "job cannot be cancelled (already completed or failed)",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "job cancellation requested",
		"job_id":  id,
	})
}

// GetStats handles GET /api/v1/jobs/stats
func (h *JobHandler) GetStats(c *gin.Context) {
	stats, err := h.manager.Stats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.JobStatsResponse{
		TotalJobs:   stats.TotalJobs,
		QueueLength: stats.QueueLength,
		MaxWorkers:  stats.MaxWorkers,
	})
}
