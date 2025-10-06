package models

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// DataListResponse represents data list response
type DataListResponse struct {
	Data []Data `json:"data"`
}

// DataResponse represents data response
type DataResponse struct {
	Data Data `json:"data"`
}
