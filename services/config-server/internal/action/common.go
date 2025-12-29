package action

// Pagination represents pagination parameters for list queries
type Pagination struct {
	Page  int
	Limit int
}

// NewPagination creates a new Pagination with validated values
func NewPagination(page, limit int) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return Pagination{
		Page:  page,
		Limit: limit,
	}
}

// ListResult represents a paginated list result
type ListResult[T any] struct {
	Items      []T
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

// NewListResult creates a new ListResult with calculated pagination metadata
func NewListResult[T any](items []T, pagination Pagination, total int) ListResult[T] {
	totalPages := total / pagination.Limit
	if total%pagination.Limit > 0 {
		totalPages++
	}
	return ListResult[T]{
		Items:      items,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
