package httpserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseJSON(t *testing.T) {
	// Given:
	given := []byte(`{"message":true}`)
	r := httptest.NewRequest(http.MethodPost, "/some/path", bytes.NewReader(given))
	reader := r.Body
	type want struct {
		Message bool `json:"message,omitempty"`
	}

	var obj want

	// When:
	err := ParseJSON(reader, &obj)

	// Then:
	require.Nil(t, err)
	require.Equal(t, true, obj.Message)
}
