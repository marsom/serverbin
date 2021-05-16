package httphandler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type origin struct {
	ClientIP string `json:"client-ip,omitempty"`
	RemoteIP string `json:"remote-ip,omitempty"`
}

type cookie struct {
	Path  string `json:"path,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Payload struct {
	Base64 string      `json:"base64,omitempty"`
	Json   interface{} `json:"json,omitempty"`
}

type multiPart struct {
	*Payload
	Name     string `json:"name,omitempty"`
	FileName string `json:"filename,omitempty"`
}

type response struct {
	Errors    []string     `json:"errors,omitempty"`
	Headers   http.Header  `json:"headers,omitempty"`
	Cookies   []cookie     `json:"cookies,omitempty"`
	Multipart []*multiPart `json:"multiPart,omitempty"`
	Form      url.Values   `json:"form,omitempty"`
	Payload   *Payload     `json:"payload,omitempty"`
	Origin    origin       `json:"origin,omitempty"`
}

func newOrigin(config Server, r *http.Request) origin {
	data := origin{}

	if remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && remoteAddr != "" {
		data.RemoteIP = remoteAddr
		data.ClientIP = remoteAddr

		// X-Forwarded-For: <client>, <proxy1>, <proxy2>
		// X-Forwarded-For: 192.0.2.43, "[2001:db8:cafe::17]"
		if header := r.Header.Get("X-Forwarded-For"); header != "" {
			if clientIP := net.ParseIP(strings.Trim(strings.Split(header, ",")[0], "\"[]")); clientIP != nil {
				if remoteIp := net.ParseIP(remoteAddr); remoteIp != nil {
					for _, network := range config.TrustedAddresses {
						if network.Contains(remoteIp) {
							data.ClientIP = clientIP.String()
							break
						}
					}
				}
			}
		}

		// Forwarded: for=192.0.2.43, for="[2001:db8:cafe::17]"
		// Forwarded: for=123.34.567.89
		// Forwarded: for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=10.1.2.3
		if header := r.Header.Get("Forwarded"); header != "" {
			for _, part := range strings.Split(header, ",") {
				part := strings.TrimSpace(part)

				if strings.HasPrefix(part, "for=") {
					if clientIP := net.ParseIP(strings.Trim(strings.Split(strings.TrimPrefix(part, "for="), ";")[0], "\"[]")); clientIP != nil {
						data.ClientIP = clientIP.String()
						break
					}
				}
			}

		}
	}

	return data
}

func newPayload(data []byte) *Payload {
	if len(data) > 0 {
		p := &Payload{
			Base64: base64.StdEncoding.EncodeToString(data),
		}

		var jsonData interface{}

		if err := json.Unmarshal(data, &jsonData); err == nil {
			p.Json = jsonData
		}

		return p
	}

	return nil
}

func newResponse(config Server, r *http.Request, errs ...error) *response {
	resp := response{
		Headers:   r.Header,
		Multipart: []*multiPart{},
		Origin:    newOrigin(config, r),
		Errors:    nil,
	}

	// cookies
	if cookies := r.Cookies(); len(cookies) > 0 {
		resp.Cookies = make([]cookie, len(cookies))

		for i, c := range cookies {
			resp.Cookies[i] = cookie{
				Path:  c.Path,
				Name:  c.Name,
				Value: c.Value,
			}
		}
	}

	// errors
	if len(errs) > 0 {
		for _, err := range errs {
			if err != nil {
				resp.Errors = append(resp.Errors, err.Error())
			}
		}
	}

	// multiPart/form-data
	reader, err := r.MultipartReader()
	if err == nil {
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("could not get next part: %s", err)
				continue
			}

			// anonymous function to call defer inside loop
			func() {
				defer func(part *multipart.Part) {
					if err := part.Close(); err != nil {
						log.Printf("failed closing multipart '%s': %s", part.FormName(), err)
					}
				}(part)

				buf := &bytes.Buffer{}

				if _, err = io.Copy(buf, part); err == nil {
					resp.Multipart = append(resp.Multipart, &multiPart{
						Payload:  newPayload(buf.Bytes()),
						Name:     part.FormName(),
						FileName: part.FileName(),
					})
				}
			}()
		}
	}

	// application/x-www-form-urlencoded
	if err := r.ParseForm(); err == nil {
		resp.Form = r.Form
	}

	// Payload
	body, err := io.ReadAll(r.Body)
	defer func(r io.Closer) {
		if err := r.Close(); err != nil {
			log.Printf("close request body failed: %s", err)
		}
	}(r.Body)

	if err == nil {
		resp.Payload = newPayload(body)
	}

	return &resp
}
