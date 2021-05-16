// Package proxyprotocol provides a experimental proxy protocol inner/outer implementation.

package proxyprotocol

import (
	"io"
	"net"
)

// ProxyProtocol gives minimal information about the detected protocol
type ProxyProtocol interface {
	Version() string
	Protocol() string
	Source() net.Addr
	Destination() net.Addr
}

// Reader with proxy protocol support
type Reader interface {
	io.Reader

	ProxyProtocol() (ProxyProtocol, bool)
	Error() error // proxy protocol related error
}

// NewReader create proxy protocol reader for the proxy protocol inner and/or outer
func NewReader(r io.Reader, v1, v2 bool) Reader {
	if v1 && v2 {
		return newReaderV1V2(r)
	}

	if v1 {
		return newReaderV1(r)
	}

	if v2 {
		return newReaderV2(r)
	}

	return &noneReader{r}
}

