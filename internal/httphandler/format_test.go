package httphandler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


type formatTest struct {
	request  request
	response response
}

var formatTests = []formatTest{
	{
		request: request{
			method: "GET",
			body:   nil,
			header: nil,
		},
		response: response{
			Errors:    nil,
			Headers:   nil,
			Cookies:   nil,
			Multipart: nil,
			Form:      nil,
			Payload:   nil,
			Origin:    origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
	{
		request: request{
			method: "PUT",
			body:   nil,
			header: nil,
		},
		response: response{
			Errors:    nil,
			Headers:   nil,
			Cookies:   nil,
			Multipart: nil,
			Form:      nil,
			Payload:   nil,
			Origin:    origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
	{
		request: request{
			method: "PUT",
			body:   strings.NewReader("test"),
			header: nil,
		},
		response: response{
			Errors:    nil,
			Headers:   nil,
			Cookies:   nil,
			Multipart: nil,
			Form:      nil,
			Payload:   &Payload{
				Base64: base64.StdEncoding.EncodeToString([]byte("test")),
				Json:   nil,
			},
			Origin:    origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
}

func TestFormat(t *testing.T) {
	for i, tt := range formatTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, "http://localhost/foo", tt.request.body)
			req.Header = tt.request.header

			w := httptest.NewRecorder()
			handler := format(Server{
				MaxRequestBody: 1024,
			}, req, 200)
			handler(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.Nil(t, err)

			r := response{}
			err = json.Unmarshal(body, &r)
			assert.Nil(t, err)

			assert.Equal(t, tt.response, r)
		})
	}
}

func TestFormat2(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost/foo", nil)
	w := httptest.NewRecorder()

	handler := format(Server{}, req, 200)

	handler(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	require.Equal(t, 200, resp.StatusCode)

	r := response{}
	err = json.Unmarshal(body, &r)
	assert.Nil(t, err)

	assert.Len(t, r.Errors, 0)
	assert.Len(t, r.Cookies, 0)
	assert.Len(t, r.Form, 0)
}
