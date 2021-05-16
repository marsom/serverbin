package httphandler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	server0 = Server{
		MaxRequestBody:    1024,
		BaseUrl:           nil,
		ManagementBaseUrl: nil,
		TrustedAddresses:  nil,
	}
)

type request struct {
	method string
	body   io.Reader
	header http.Header
}

type responseTest struct {
	Server   Server
	request  request
	response response
}


var responseTests = []responseTest{
	{
		Server: Server{
			MaxRequestBody:    32,
		},
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
			Origin: origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
	{
		Server: Server{
			MaxRequestBody:    32,
		},
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
			Origin: origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
	{
		Server: Server{
			MaxRequestBody:    32,
		},
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
			Payload: &Payload{
				Base64: base64.StdEncoding.EncodeToString([]byte("test")),
				Json:   nil,
			},
			Origin: origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
	{
		Server: Server{
			MaxRequestBody:    1024,
		},
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
			Payload: &Payload{
				Base64: base64.StdEncoding.EncodeToString([]byte("test")),
				Json:   nil,
			},
			Origin: origin{
				ClientIP: "192.0.2.1",
				RemoteIP: "192.0.2.1",
			},
		},
	},
}

func TestResponse(t *testing.T) {
	for i, tt := range responseTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, "http://localhost/foo", tt.request.body)
			req.Header = tt.request.header

			w := httptest.NewRecorder()
			handler := format(tt.Server, req, 200)
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

func TestMultipartRequest(t *testing.T) {
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)

	fw, err := mw.CreateFormField("a")
	assert.Nil(t, err)
	_, _ = io.Copy(fw, strings.NewReader("test1"))

	fw, err = mw.CreateFormField("b")
	assert.Nil(t, err)
	_, _ = io.Copy(fw, strings.NewReader("test2"))

	mw.Close()

	req := httptest.NewRequest("PUT", "http://localhost/foo", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	w := httptest.NewRecorder()
	handler := format(server0, req, 200)
	handler(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	r := response{}
	err = json.Unmarshal(body, &r)
	assert.Nil(t, err)

	expected := response{
		Errors:    nil,
		Headers:   http.Header{
			"Content-Type": []string{mw.FormDataContentType()},
		},
		Cookies:   nil,
		Multipart: []*multiPart{
			{
				Payload: &Payload{
					Base64: base64.StdEncoding.EncodeToString([]byte("test1")),
					Json:   nil,
				},
				Name:     "a",
				FileName: "",
			},
			{
				Payload: &Payload{
					Base64: base64.StdEncoding.EncodeToString([]byte("test2")),
					Json:   nil,
				},
				Name:     "b",
				FileName: "",
			},
		},
		Form:      nil,
		Payload:   nil,
		Origin: origin{
			ClientIP: "192.0.2.1",
			RemoteIP: "192.0.2.1",
		},
	}

	assert.Equal(t, expected, r)
}
