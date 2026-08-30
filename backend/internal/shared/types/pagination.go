package types

// PaginationRequest represents pagination parameters from query string.
type PaginationRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir"` // asc | desc
}

// Normalize applies defaults and clamps values.
func (p *PaginationRequest) Normalize(maxPageSize int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
	if p.SortDir != "asc" && p.SortDir != "desc" {
		p.SortDir = "desc"
	}
}

// Offset returns the SQL offset value.
func (p *PaginationRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PaginationMeta holds pagination metadata returned in API responses.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationMeta calculates total pages from total records.
func NewPaginationMeta(page, pageSize int, total int64) PaginationMeta {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
