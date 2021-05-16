package httphandler

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost/foo", nil)
	w := httptest.NewRecorder()

	handler := status(Server{}, 200)
	handler(w, req)

	resp := w.Result()
	_, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	require.Equal(t, 200, resp.StatusCode)
}
