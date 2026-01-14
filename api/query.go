package api

import (
	"net/http"
	"strconv"
)

// QueryParams provides helpers for parsing URL query parameters
type QueryParams struct {
	r *http.Request
}

// Query returns a QueryParams helper for the request
func Query(r *http.Request) QueryParams {
	return QueryParams{r: r}
}

// String returns a query parameter as string, or default if not present
func (q QueryParams) String(key, def string) string {
	if v := q.r.URL.Query().Get(key); v != "" {
		return v
	}
	return def
}

// Int returns a query parameter as int, or default if not present/invalid
func (q QueryParams) Int(key string, def int) int {
	v := q.r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

// Bool returns a query parameter as bool, or default if not present
func (q QueryParams) Bool(key string, def bool) bool {
	v := q.r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// Pagination is common pagination parameters
type Pagination struct {
	Page    int
	PerPage int
	Offset  int
}

// DefaultPagination is the default pagination settings
var DefaultPagination = Pagination{Page: 1, PerPage: 20}

// ParsePagination extracts pagination from query params
func (q QueryParams) Pagination() Pagination {
	p := Pagination{
		Page:    q.Int("page", DefaultPagination.Page),
		PerPage: q.Int("per_page", DefaultPagination.PerPage),
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = DefaultPagination.PerPage
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	p.Offset = (p.Page - 1) * p.PerPage
	return p
}

// PaginatedResult wraps a slice with pagination metadata
type PaginatedResult[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPaginatedResult creates a paginated result from a slice
func NewPaginatedResult[T any](data []T, p Pagination, total int) PaginatedResult[T] {
	totalPages := total / p.PerPage
	if total%p.PerPage > 0 {
		totalPages++
	}
	return PaginatedResult[T]{
		Data:       data,
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
