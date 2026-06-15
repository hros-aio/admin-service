// Package response provides standard HTTP response formats.
package response

// SuccessResponse is the standard success response wrapper.
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

// NewSuccessResponse creates a new SuccessResponse.
func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{
		Data: data,
	}
}
