package models

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type APIInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Description string `json:"description"`
}

func NewPaginatedResponse(data interface{}, total, limit, offset int) PaginatedResponse {
	return PaginatedResponse{
		Data:    data,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+limit < total,
	}
}
