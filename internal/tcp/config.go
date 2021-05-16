package tcp

import "net"

type Server struct {
	MaxBufferSize    int64
	TrustedAddresses []*net.IPNet
}

type Config struct {
	Server Server
}
