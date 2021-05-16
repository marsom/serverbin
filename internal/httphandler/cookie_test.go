package httphandler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookieHanlderPUT(t *testing.T) {
	handler := &cookieHandler{
		Server: Server{},
		Cookie: Cookie{
			Names:    []string{"a"},
			HttpOnly: false,
			Secure:   true,
		},
		Path: "/",
	}

	req := httptest.NewRequest("PUT", "http://localhost/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{`a="test /"; Path=/; Secure; SameSite=Strict`}, resp.Header.Values("Set-Cookie"))

	r := response{}
	err = json.Unmarshal(body, &r)
	assert.Nil(t, err)
}

func TestCookieHanlderDELETE(t *testing.T) {
	handler := &cookieHandler{
		Server: Server{},
		Cookie: Cookie{
			Names:    []string{"a"},
			HttpOnly: false,
			Secure:   true,
		},
		Path: "/",
	}

	req := httptest.NewRequest("DELETE", "http://localhost/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, []string{"a=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}, resp.Header.Values("Set-Cookie"))

	r := response{}
	err = json.Unmarshal(body, &r)
	assert.Nil(t, err)
}

func TestCookieHanlderGET(t *testing.T) {
	handler := &cookieHandler{
		Server: Server{},
		Cookie: Cookie{
			Names:    []string{"a"},
			HttpOnly: false,
			Secure:   true,
		},
		Path: "/",
	}

	req := httptest.NewRequest("GET", "http://localhost/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	require.Equal(t, 200, resp.StatusCode)

	r := response{}
	err = json.Unmarshal(body, &r)
	assert.Nil(t, err)
}
