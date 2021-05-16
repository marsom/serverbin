package proxyprotocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// IPs
	srcIPv4 = net.ParseIP("127.0.0.1").To4()
	srcIPv6 = net.ParseIP("::1").To16()
	dstIPv4 = net.ParseIP("127.0.0.2").To4()
	dstIPv6 = net.ParseIP("::2").To16()

	// TCP
	srcTCPv4 = &net.TCPAddr{IP: srcIPv4, Port: 50000}
	srcTCPv6 = &net.TCPAddr{IP: srcIPv6, Port: 50000}
	dstTCPv4 = &net.TCPAddr{IP: dstIPv4, Port: 8080}
	dstTCPv6 = &net.TCPAddr{IP: dstIPv6, Port: 7070}

	// UDP
	srcUDPv4 = &net.UDPAddr{IP: srcIPv4, Port: 50000}
	srcUDPv6 = &net.UDPAddr{IP: srcIPv6, Port: 50000}
	dstUDPv4 = &net.UDPAddr{IP: dstIPv4, Port: 8080}
	dstUDPv6 = &net.UDPAddr{IP: dstIPv6, Port: 7070}

	// UNIX
	unixStreamAddr   = &net.UnixAddr{Net: "unix", Name: "/path/to/unix.sock"}
	unixDatagramAddr = &net.UnixAddr{Net: "unixgram", Name: "/path/to/unix.sock"}
)

func asPort(addr net.Addr) []byte {
	var data []byte

	switch v := addr.(type) {
	case *net.TCPAddr:
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(v.Port))
		data = append(data, portBytes...)
	case *net.UDPAddr:
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(v.Port))
		data = append(data, portBytes...)
	}

	return data
}

func asAddr(addr net.Addr) []byte {
	var data []byte

	switch v := addr.(type) {
	case *net.TCPAddr:
		if ip := v.IP.To4(); ip != nil {
			data = append(data, ip...)
		} else if ip := v.IP.To16(); ip != nil {
			data = append(data, ip...)
		}
	case *net.UDPAddr:
		if ip := v.IP.To4(); ip != nil {
			data = append(data, ip...)
		} else if ip := v.IP.To16(); ip != nil {
			data = append(data, ip...)
		}
	case *net.UnixAddr:
		fixedSize := make([]byte, 108)
		copy(fixedSize, v.Name)

		data = append(data, fixedSize...)
	}

	return data
}

func asSrcDstAddr(src net.Addr, dst net.Addr) []byte {
	var data []byte

	data = append(data, asAddr(src)...)
	data = append(data, asAddr(dst)...)
	data = append(data, asPort(src)...)
	data = append(data, asPort(dst)...)

	return data
}

func asLenghtV2(length uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, length)
	return b
}

var v2Tests = []struct {
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
		output: "PROXY UNKNOWN\r\n",
	},
	{
		input:  "PROXY UNKNOWN kjk jkj\r\nGAGA",
		output: "PROXY UNKNOWN kjk jkj\r\nGAGA",
	},
	{
		input:  string(v2Signature),
		output: string(v2Signature),
	},
	{
		input: string(
			append(
				v2Signature,
				v2ProtocolVersionAndCommandLocal,
				v2TransportProtocolAndAddressFamilyUNSPEC,
				asLenghtV2(0)[0], asLenghtV2(0)[1]),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolUNSPEC,
			src:      nil,
			dst:      nil,
		},
	},
	{
		input: string(
			append(append(
				v2Signature,
				v2ProtocolVersionAndCommandProxy,
				v2TransportProtocolAndAddressFamilyTCPv4,
				asLenghtV2(12)[0], asLenghtV2(12)[1],
			), asSrcDstAddr(srcTCPv4, dstTCPv4)...),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolTCPv4,
			src:      srcTCPv4,
			dst:      dstTCPv4,
		},
	},
	{
		input: string(
			append(append(
				v2Signature,
				v2ProtocolVersionAndCommandProxy,
				v2TransportProtocolAndAddressFamilyUDPv4,
				asLenghtV2(12)[0], asLenghtV2(12)[1],
			), asSrcDstAddr(srcUDPv4, dstUDPv4)...),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolUDPv4,
			src:      srcUDPv4,
			dst:      dstUDPv4,
		},
	},
	{
		input: string(
			append(append(
				v2Signature,
				v2ProtocolVersionAndCommandProxy,
				v2TransportProtocolAndAddressFamilyTCPv6,
				asLenghtV2(36)[0], asLenghtV2(36)[1],
			), asSrcDstAddr(srcTCPv6, dstTCPv6)...),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolTCPv6,
			src:      srcTCPv6,
			dst:      dstTCPv6,
		},
	},
	{
		input: string(
			append(append(
				v2Signature,
				v2ProtocolVersionAndCommandProxy,
				v2TransportProtocolAndAddressFamilyUDPv6,
				asLenghtV2(36)[0], asLenghtV2(36)[1],
			), asSrcDstAddr(srcUDPv6, dstUDPv6)...),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolUDPv6,
			src:      srcUDPv6,
			dst:      dstUDPv6,
		},
	},
	{
		input: string(
			append(append(
				v2Signature,
				v2ProtocolVersionAndCommandProxy,
				v2TransportProtocolAndAddressFamilyUnixStream,
				asLenghtV2(216)[0], asLenghtV2(216)[1],
			), asSrcDstAddr(unixStreamAddr, unixStreamAddr)...),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolUnixStream,
			src:      unixStreamAddr,
			dst:      unixStreamAddr,
		},
	},
	{
		input: string(
			append(append(
				v2Signature,
				v2ProtocolVersionAndCommandProxy,
				v2TransportProtocolAndAddressFamilyUnixDatagram,
				asLenghtV2(216)[0], asLenghtV2(216)[1],
			), asSrcDstAddr(unixDatagramAddr, unixDatagramAddr)...),
		),
		output: "",
		protocol: &v2{
			protocol: v2ProtocolUnixDatagram,
			src:      unixDatagramAddr,
			dst:      unixDatagramAddr,
		},
	},
}

func TestReaderV2(t *testing.T) {
	for i, tt := range v2Tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			r := newReaderV2(bytes.NewReader([]byte(tt.input)))
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
