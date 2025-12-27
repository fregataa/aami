package handler

import (
	"errors"
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// respondError sends an error response with appropriate status code
func respondError(c *gin.Context, err error) {
	var statusCode int
	var code string

	switch {
	case errors.Is(err, service.ErrNotFound):
		statusCode = http.StatusNotFound
		code = "NOT_FOUND"
	case errors.Is(err, service.ErrAlreadyExists):
		statusCode = http.StatusConflict
		code = "ALREADY_EXISTS"
	case errors.Is(err, service.ErrForeignKeyViolation):
		statusCode = http.StatusBadRequest
		code = "FOREIGN_KEY_VIOLATION"
	case errors.Is(err, service.ErrCircularReference):
		statusCode = http.StatusBadRequest
		code = "CIRCULAR_REFERENCE"
	case errors.Is(err, service.ErrInUse):
		statusCode = http.StatusConflict
		code = "IN_USE"
	case errors.Is(err, service.ErrInvalidInput):
		statusCode = http.StatusBadRequest
		code = "INVALID_INPUT"
	default:
		// Check if it's a validation error
		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			statusCode = http.StatusBadRequest
			code = "VALIDATION_ERROR"
			c.JSON(statusCode, dto.ErrorResponse{
				Error: err.Error(),
				Code:  code,
				Details: map[string]interface{}{
					"field": validationErr.Field,
				},
			})
			return
		}

		// Internal server error for unknown errors
		statusCode = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
	}

	c.JSON(statusCode, dto.ErrorResponse{
		Error: err.Error(),
		Code:  code,
	})
}

// getPagination extracts pagination parameters from query string
func getPagination(c *gin.Context) dto.PaginationRequest {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		// Return default pagination if binding fails
		return dto.DefaultPagination()
	}
	pagination.Normalize()
	return pagination
}

// respondList sends a paginated list response
func respondList(c *gin.Context, data interface{}, total int, pagination dto.PaginationRequest) {
	c.JSON(http.StatusOK, dto.ListResponse{
		Data:       data,
		Pagination: dto.NewPaginationResponse(pagination, total),
	})
}
