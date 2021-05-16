package httphandler

import (
	"net"
	"net/url"
)

type Server struct {
	MaxRequestBody    int64
	BaseUrl           *url.URL
	ManagementBaseUrl *url.URL
	TrustedAddresses  []*net.IPNet
}

type Config struct {
	Path     string
	Server   Server
	Cookie   *Cookie
	Delay    *Delay
	Slow     *Slow
	Redirect *Redirect
}
