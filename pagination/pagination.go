package pagination

// PaginatedDTO wraps a paginated list response with total count metadata.
type PaginatedDTO[T any] struct {
	Total int `json:"total"`
	Data  []T `json:"data"`
}

// Paginate creates a PaginatedDTO with the given total count and data slice.
func Paginate[T any](total int, data []T) PaginatedDTO[T] {
	return PaginatedDTO[T]{
		Total: total,
		Data:  data,
	}
}
