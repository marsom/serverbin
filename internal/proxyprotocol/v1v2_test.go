// Package proxyprotocol provides a experimental proxy protocol inner/outer implementation.

package proxyprotocol

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV1V2WithV1Data(t *testing.T) {
	for i, tt := range v1Tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.input), func(t *testing.T) {
			r := NewReader(bytes.NewReader([]byte(tt.input)), true, true)
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