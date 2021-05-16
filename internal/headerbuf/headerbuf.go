// Package headerbuf provides a reader implementation which buffers the head of
// another reader, and delegates it's work to the underlying reader if the buffer
// is consumed.
// this could easy achieved by the bufio package.

package headerbuf

import (
	"errors"
	"io"
)

// NewReader create a new reader which buffers the head of another reader
func NewReader(rd io.Reader) *Reader {
	return &Reader{
		buf: nil,
		rd:  rd,
		err: nil,
	}
}

// Reader a reader which buffers the head of another reader
type Reader struct {
	buf []byte
	rd  io.Reader
	err error
}

// Peek read n bytes without consuming them
func (r *Reader) Peek(n int) ([]byte, error) {
	p := make([]byte, n)

	n, err := r.rd.Read(p)

	r.err = err
	r.buf = append(r.buf, p[0:n]...)

	// maybe we should copy the buffer because the user could change it
	// but for our case it's enough, we know we do not modify it

	return p[0:n], err
}

// PeekByte read one byte without consuming it
func (r *Reader) PeekByte() (byte, error) {
	p, err := r.Peek(1)

	if len(p) > 0 {
		return p[0], err
	}

	return ' ', err
}

// Done consumes given length from the buffer, if n < 0 everything will be consumed
func (r *Reader) Done(n int) error {
	if bl := len(r.buf); bl > 0 {
		if n == bl || n < 0 {
			// consume everything in the buffer
			r.buf = nil
		} else if n < bl {
			// consume the header
			r.buf = r.buf[n:bl]
		} else {
			// you try to consume more bytes form the buffer than actually exits
			// this is probably an codding issue
			return errors.New("header is bigger than the buffer")
		}
	}

	return nil
}

// Read reads from the head buffer or delegate the work to the underlying reader
func (r *Reader) Read(p []byte) (n int, err error) {
	if bl := len(r.buf); bl > 0 {
		pl := len(p)

		//   p: <------->
		// buf: <------->
		if pl == bl {
			copy(p, r.buf)
			r.buf = nil

			return pl, r.err
		}

		//   p: <-->
		// buf: <------->
		if pl < bl {
			copy(p, r.buf[0:pl])
			r.buf = r.buf[pl:]

			return pl, nil
		}

		//   p: <------------->
		// buf: <------->
		if pl > bl {
			buffer := make([]byte, pl-bl)

			n, err := r.rd.Read(buffer)

			copy(p, append(r.buf, buffer[0:n]...))

			r.buf = nil

			if n < len(buffer) {
				return bl + n, io.EOF
			}

			return bl + n, err
		}
	}

	// delegate to underlying reader
	n, err = r.rd.Read(p)
	if n < len(p) {
		return n, io.EOF
	}

	return n, err
}
