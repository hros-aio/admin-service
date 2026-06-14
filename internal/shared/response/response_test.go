package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"foo": "bar"}
	resp := NewSuccessResponse(data)
	assert.Equal(t, data, resp.Data)
}
