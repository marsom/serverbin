package proxyprotocol

import (
	"io"
)

// Fulfills the Reader interface but does no proxy protocol detection
type noneReader struct {
	r io.Reader
}

func (rd *noneReader) Read(p []byte) (n int, err error) {
	return rd.r.Read(p)
}

func (rd *noneReader) ProxyProtocol() (ProxyProtocol, bool) {
	return nil, false
}

func (rd *noneReader) Error() error {
	return nil
}

