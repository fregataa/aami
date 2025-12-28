package handler

import (
	"errors"
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/gin-gonic/gin"
)

// respondError sends an error response with appropriate status code
func respondError(c *gin.Context, err error) {
	var statusCode int
	var code string

	switch {
	// Resource errors
	case errors.Is(err, domainerrors.ErrNotFound):
		statusCode = http.StatusNotFound
		code = "NOT_FOUND"
	case errors.Is(err, domainerrors.ErrAlreadyExists):
		statusCode = http.StatusConflict
		code = "ALREADY_EXISTS"
	case errors.Is(err, domainerrors.ErrInUse):
		statusCode = http.StatusConflict
		code = "IN_USE"

	// Database constraint errors
	case errors.Is(err, domainerrors.ErrDuplicateKey):
		statusCode = http.StatusConflict
		code = "DUPLICATE_KEY"
	case errors.Is(err, domainerrors.ErrForeignKeyViolation):
		statusCode = http.StatusBadRequest
		code = "FOREIGN_KEY_VIOLATION"
	case errors.Is(err, domainerrors.ErrConstraintViolation):
		statusCode = http.StatusBadRequest
		code = "CONSTRAINT_VIOLATION"

	// Validation and business logic errors
	case errors.Is(err, domainerrors.ErrInvalidInput):
		statusCode = http.StatusBadRequest
		code = "INVALID_INPUT"
	case errors.Is(err, domainerrors.ErrCircularReference):
		statusCode = http.StatusBadRequest
		code = "CIRCULAR_REFERENCE"
	case errors.Is(err, domainerrors.ErrCannotRemoveLastGroup):
		statusCode = http.StatusBadRequest
		code = "CANNOT_REMOVE_LAST_GROUP"

	// Bootstrap token errors
	case errors.Is(err, domainerrors.ErrTokenExpired):
		statusCode = http.StatusBadRequest
		code = "TOKEN_EXPIRED"
	case errors.Is(err, domainerrors.ErrTokenExhausted):
		statusCode = http.StatusBadRequest
		code = "TOKEN_EXHAUSTED"
	case errors.Is(err, domainerrors.ErrInvalidToken):
		statusCode = http.StatusBadRequest
		code = "INVALID_TOKEN"

	// Request binding/parsing errors
	case errors.Is(err, domainerrors.ErrBindingFailed):
		statusCode = http.StatusBadRequest
		code = "BINDING_FAILED"

	default:
		// Check for structured error types
		var bindingErr *domainerrors.BindingError
		if errors.As(err, &bindingErr) {
			statusCode = http.StatusBadRequest
			code = "BINDING_ERROR"
			c.JSON(statusCode, dto.ErrorResponse{
				Error: bindingErr.Message,
				Code:  code,
			})
			return
		}

		var validationErr *domainerrors.ValidationError
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

		var dbErr *domainerrors.DatabaseError
		if errors.As(err, &dbErr) {
			statusCode = http.StatusInternalServerError
			code = "DATABASE_ERROR"
			c.JSON(statusCode, dto.ErrorResponse{
				Error: "Database operation failed",
				Code:  code,
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
