package model

// ListResponse wraps paginated query results.
type ListResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
	Limit int `json:"limit,omitempty"`
	Page  int `json:"page,omitempty"`
}

// ErrorResponse represents a standardized API error payload.
type ErrorResponse struct {
	Detail string `json:"detail"`
	Code   string `json:"code,omitempty"`
}

// UpdatePayload wraps where-clause criteria and updated data fields.
type UpdatePayload[W, D any] struct {
	Where W `json:"where"`
	Data  D `json:"data"`
}
