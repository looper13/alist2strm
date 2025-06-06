package utils

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageData 分页数据结构
type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    PageData `json:"data"`
}

// NewResponse 创建标准响应
func NewResponse(code int, message string, data interface{}) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) Response {
	return NewResponse(200, "success", data)
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) Response {
	return NewResponse(code, message, nil)
}

// NewPageResponse 创建分页响应
func NewPageResponse(list interface{}, total int64, page, size int) PageResponse {
	return PageResponse{
		Code:    200,
		Message: "success",
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	}
}
