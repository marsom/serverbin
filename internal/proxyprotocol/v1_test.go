package proxyprotocol

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var v1Tests = []struct {
	input    string
	output   string
	protocol ProxyProtocol
}{
	{
		input:  "",
		output: "",
	},
	{
		input:  "TEST",
		output: "TEST",
	},
	{
		input:  "PROXY",
		output: "PROXY",
	},
	{
		input:  "PROXY\r\n",
		output: "PROXY\r\n",
	},
	{
		input:  "PROXY UNKNOWN\r\n",
		output: "",
	},
	{
		input:  "PROXY UNKNOWN kjk jkj\r\nGAGA",
		output: "GAGA",
		protocol: &v1{
			protocol: "UNKNOWN",
			src:      nil,
			dst:      nil,
		},
	},
}

func TestReaderV1(t *testing.T) {
	for i, tt := range v1Tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.input), func(t *testing.T) {
			r := newReaderV1(bytes.NewReader([]byte(tt.input)))
			assert.NotNil(t, r)

			b, err := io.ReadAll(r)
			assert.Nil(t, err)

			assert.Equal(t, tt.output, string(b))

			if tt.protocol != nil {
				protocol, ok := r.ProxyProtocol()

				assert.True(t, ok)

				if tt.protocol != protocol {
					assert.Equal(t, tt.protocol, protocol)
				}
			}

		})
	}
}
