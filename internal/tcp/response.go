package tcp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/marsom/serverbin/internal/proxyprotocol"
)

type origin struct {
	ClientIP      string         `json:"client-ip,omitempty"`
	RemoteIP      string         `json:"remote-ip,omitempty"`
	ProxyProtocol *proxyProtocol `json:"proxy-protocol,omitempty"`
}

type httpPayload struct {
	Method string      `json:"method,omitempty"`
	URL    string      `json:"url,omitempty"`
	Header http.Header `json:"headers,omitempty"`
}

type payload struct {
	Base64 string       `json:"base64,omitempty"`
	Json   interface{}  `json:"json,omitempty"`
	Http   *httpPayload `json:"http,omitempty"`
}

type response struct {
	Errors  []string `json:"errors,omitempty"`
	Payload *payload `json:"payload,omitempty"`
	Origin  origin   `json:"origin,omitempty"`
}

type proxyProtocol struct {
	Version     string `json:"version,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
}

func newProxyProtocol(protocol proxyprotocol.ProxyProtocol) *proxyProtocol {
	src := ""
	dst := ""

	if v := protocol.Source(); v != nil {
		src = v.String()
	}

	if v := protocol.Destination(); v != nil {
		dst = v.String()
	}

	return &proxyProtocol{
		Version:     protocol.Version(),
		Protocol:    protocol.Protocol(),
		Source:      src,
		Destination: dst,
	}
}

func newOrigin(config Server, conn net.Conn, r proxyprotocol.Reader) origin {
	data := origin{}

	if remoteAddr, _, err := net.SplitHostPort(conn.RemoteAddr().String()); err == nil && remoteAddr != "" {
		data.RemoteIP = remoteAddr
		data.ClientIP = remoteAddr
	}

	if protocol, ok := r.ProxyProtocol(); ok {
		data.ProxyProtocol = newProxyProtocol(protocol)

		// update client ip if we trust the remote ip
		if remoteIP := net.ParseIP(data.RemoteIP); remoteIP != nil {
			for _, network := range config.TrustedAddresses {
				if network.Contains(remoteIP) {
					data.ClientIP = data.RemoteIP
					break
				}
			}
		}
	}

	return data
}

func newPayload(data []byte) *payload {
	if len(data) > 0 {
		p := &payload{
			Base64: base64.StdEncoding.EncodeToString(data),
		}

		var jsonData interface{}

		if err := json.Unmarshal(data, &jsonData); err == nil {
			p.Json = jsonData
		}

		// HTTP/1.x
		if r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data))); err == nil {

			p.Http = &httpPayload{
				Method: r.Method,
				Header: r.Header,
			}

			if r.URL != nil {
				p.Http.URL = r.URL.String()
			}
		}

		return p
	}

	return nil
}

func newResponse(config Config, conn net.Conn, errs ...error) *response {
	resp := response{}

	// Read the incoming connection into the buffer.
	buffer := make([]byte, config.Server.MaxBufferSize)

	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		resp.Errors = append(resp.Errors, err.Error())
	}

	// errors
	if len(errs) > 0 {
		for _, err := range errs {
			if err != nil {
				resp.Errors = append(resp.Errors, err.Error())
			}
		}
	}

	r := proxyprotocol.NewReader(bytes.NewReader(buffer[:n]), true, false)

	body, err := io.ReadAll(r)
	if err != nil {
		resp.Errors = append(resp.Errors, err.Error())
	}

	// payload
	resp.Payload = newPayload(body)
	resp.Origin = newOrigin(config.Server, conn, r)

	return &resp
}
