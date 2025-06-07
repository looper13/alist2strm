package utils

// PaginationQuery 分页查询参数
type PaginationQuery struct {
	Page int `form:"page" binding:"required,min=1"` // 当前页码
	Size int `form:"size" binding:"required,min=1"` // 每页大小
}

// GetOffset 获取偏移量
func (p *PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.Size
}

// GetLimit 获取限制数量
func (p *PaginationQuery) GetLimit() int {
	return p.Size
}
