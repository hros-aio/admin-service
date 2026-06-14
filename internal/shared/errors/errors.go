// Package errors provides shared error types and utilities.
package errors

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// NewErrorResponse creates a new ErrorResponse.
func NewErrorResponse(code, message string, details interface{}, traceID string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
		TraceID: traceID,
	}
}
