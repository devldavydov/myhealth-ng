package entity

const (
	DefaultPage     = 1
	DefaultPageSize = 20
)

type PageRequest struct {
	Query    string
	Page     int
	PageSize int
}

type Page[T any] struct {
	Items      []T
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}
