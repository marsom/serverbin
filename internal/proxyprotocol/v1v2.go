// Package proxyprotocol provides a experimental proxy protocol v1/v2 implementation.

package proxyprotocol

import (
	"io"
)

func newReaderV1V2(r io.Reader) Reader {
	r1 := newReaderV1(r)
	r2 := newReaderV2(r1)

	return &v1v2Reader{
		inner: r1,
		outer: r2,
	}
}

// Fulfills the Reader interface but does no proxy protocol detection
type v1v2Reader struct {
	inner Reader
	outer Reader
}

func (rd *v1v2Reader) Read(p []byte) (n int, err error) {
	return rd.outer.Read(p)
}

func (rd *v1v2Reader) ProxyProtocol() (ProxyProtocol, bool) {
	if protocol, ok := rd.outer.ProxyProtocol(); ok {
		return protocol, ok
	}
	if protocol, ok := rd.inner.ProxyProtocol(); ok {
		return protocol, ok
	}

	return nil, false
}

func (rd *v1v2Reader) Error() error {
	if _, ok := rd.outer.ProxyProtocol(); ok {
		return rd.outer.Error()
	}
	if _, ok := rd.inner.ProxyProtocol(); ok {
		return rd.inner.Error()
	}

	return nil
}

