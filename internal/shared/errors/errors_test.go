package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse("test_code", "test_message", "test_details", "test_trace")
	assert.Equal(t, "test_code", resp.Code)
	assert.Equal(t, "test_message", resp.Message)
	assert.Equal(t, "test_details", resp.Details)
	assert.Equal(t, "test_trace", resp.TraceID)
}
