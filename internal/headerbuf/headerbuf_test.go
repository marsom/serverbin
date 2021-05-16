package headerbuf

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	input    string
	peek     int
	consume  int
	output   string
}

func (t test) String() string {
	return fmt.Sprintf("input:%s,peek:%d,consume:%d,output:%s", t.input, t.peek, t.consume, t.output)
}

var peekAndConsumeTests = []test{
	{input: "0123456789", peek: 10, consume: 0, output: "0123456789"},
	{input: "0123456789", peek: 10, consume: 10, output: ""},
	{input: "0123456789", peek: 10, consume: 5, output: "56789"},
	{input: "0123456789", peek: 5, consume: 0, output: "0123456789"},
	{input: "0123456789", peek: 5, consume: 5, output: "56789"},
	{input: "0123456789", peek: 5, consume: 4, output: "456789"},
	{input: "0123456789", peek: 20, consume: 0, output: "0123456789"},
	{input: "0123456789", peek: 20, consume: 5, output: "56789"},
	{input: "", peek: 20, consume: 0, output: ""},
}

func TestPeekAndConsume(t *testing.T) {
	for _, tt := range peekAndConsumeTests {
		t.Run(tt.String(), func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			_,_ = r.Peek(tt.peek)
			_ = r.Done(tt.consume)

			output, err := io.ReadAll(r)
			assert.Nil(t, err)
			assert.Equal(t, tt.output, string(output))
		})
	}
}

func TestReadEqual(t *testing.T) {
	for _, tt := range peekAndConsumeTests {
		t.Run(tt.String(), func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			_,_ = r.Peek(tt.peek)
			_ = r.Done(tt.consume)

			if l := len(tt.output); l > 0 {
				buf := make([]byte, l)

				n, err := r.Read(buf)
				assert.Nil(t, err)
				assert.Equal(t, l, n)
				assert.Equal(t, []byte(tt.output), buf)

				n, err = r.Read(buf)
				assert.Equal(t, io.EOF, err)
				assert.Equal(t, 0, n)
			}
		})
	}
}

func TestReadLess(t *testing.T) {
	for _, tt := range peekAndConsumeTests {
		t.Run(tt.String(), func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			_,_ = r.Peek(tt.peek)
			_ = r.Done(tt.consume)

			if l := len(tt.output); l > 2 {
				buf := make([]byte, 2)

				n, err := r.Read(buf)
				assert.Nil(t, err)
				assert.Equal(t, 2, n)
				assert.Equal(t, []byte(tt.output[:2]), buf)
			}
		})
	}
}

func TestReadMore(t *testing.T) {
	for _, tt := range peekAndConsumeTests {
		t.Run(tt.String(), func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			_,_ = r.Peek(tt.peek)
			_ = r.Done(tt.consume)


			buf := make([]byte, len(tt.output) + 10)

			n, err := r.Read(buf)
			assert.Equal(t, io.EOF, err)
			assert.Equal(t, len(tt.output), n)
			assert.Equal(t, []byte(tt.output), buf[0:n])
		})
	}
}

func TestPeek(t *testing.T)   {
	r := NewReader(strings.NewReader("0123456789"))

	buf, err := r.Peek(3)
	assert.Nil(t, err)
	assert.Equal(t, []byte("012"), buf)
	assert.Equal(t, []byte("012"), r.buf)

	buf, err = r.Peek(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte{}, buf)
	assert.Equal(t, []byte("012"), r.buf)

	b, err := r.PeekByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('3'), b)
	assert.Equal(t, []byte("0123"), r.buf)

	buf, err = r.Peek(6)
	assert.Nil(t, err)
	assert.Equal(t, []byte("456789"), buf)
	assert.Equal(t, []byte("0123456789"), r.buf)

	buf, err = r.Peek(6)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, []byte{}, buf)
}
