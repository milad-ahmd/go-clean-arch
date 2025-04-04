package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
)

// Response is the standard API response structure
type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Errors     interface{} `json:"errors,omitempty"`
	StatusCode int         `json:"status_code"`
	Timestamp  time.Time   `json:"timestamp"`
}

// Meta contains pagination metadata
type Meta struct {
	Page      int `json:"page"`
	PerPage   int `json:"per_page"`
	TotalPage int `json:"total_page"`
	Total     int `json:"total"`
}

// PaginatedResponse is a response with pagination
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Meta       Meta        `json:"meta"`
	StatusCode int         `json:"status_code"`
	Timestamp  time.Time   `json:"timestamp"`
}

// NewResponse creates a new response
func NewResponse(success bool, message string, data interface{}, statusCode int) *Response {
	return &Response{
		Success:    success,
		Message:    message,
		Data:       data,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, err error, statusCode int) *Response {
	var errorData interface{}
	
	// Handle different error types
	switch e := err.(type) {
	case *errors.ValidationError:
		errorData = e.Errors
	case *errors.AppError:
		errorData = e.Error()
	default:
		errorData = err.Error()
	}

	return &Response{
		Success:    false,
		Message:    message,
		Errors:     errorData,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(message string, data interface{}, page, perPage, total int, statusCode int) *PaginatedResponse {
	totalPage := total / perPage
	if total%perPage > 0 {
		totalPage++
	}

	return &PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Meta: Meta{
			Page:      page,
			PerPage:   perPage,
			TotalPage: totalPage,
			Total:     total,
		},
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Success sends a success response
func Success(w http.ResponseWriter, message string, data interface{}, statusCode int) {
	resp := NewResponse(true, message, data, statusCode)
	JSON(w, statusCode, resp)
}

// Error sends an error response
func Error(w http.ResponseWriter, message string, err error, statusCode int) {
	resp := NewErrorResponse(message, err, statusCode)
	JSON(w, statusCode, resp)
}

// Paginated sends a paginated response
func Paginated(w http.ResponseWriter, message string, data interface{}, page, perPage, total int, statusCode int) {
	resp := NewPaginatedResponse(message, data, page, perPage, total, statusCode)
	JSON(w, statusCode, resp)
}
