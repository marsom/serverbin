package proxyprotocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/marsom/serverbin/internal/headerbuf"
)

// Signature
var (
	v2Signature = []byte{'\x0D', '\x0A', '\x0D', '\x0A', '\x00', '\x0D', '\x0A', '\x51', '\x55', '\x49', '\x54', '\x0A'}
)

// The next byte (the 13th one) is the protocol version and command.
const (
	v2ProtocolVersionAndCommandLocal byte = '\x20'
	v2ProtocolVersionAndCommandProxy byte = '\x21'
)

// The 14th byte contains the transport protocol and address family. The highest 4
// bits contain the address family, the lowest 4 bits contain the protocol.
const (
	v2TransportProtocolAndAddressFamilyUNSPEC       byte = '\x00'
	v2TransportProtocolAndAddressFamilyTCPv4        byte = '\x11'
	v2TransportProtocolAndAddressFamilyUDPv4        byte = '\x12'
	v2TransportProtocolAndAddressFamilyTCPv6        byte = '\x21'
	v2TransportProtocolAndAddressFamilyUDPv6        byte = '\x22'
	v2TransportProtocolAndAddressFamilyUnixStream   byte = '\x31'
	v2TransportProtocolAndAddressFamilyUnixDatagram byte = '\x32'

	v2ProtocolUNSPEC       = "UNSPEC"
	v2ProtocolTCPv4        = "TCPv4"
	v2ProtocolUDPv4        = "UDPv4"
	v2ProtocolTCPv6        = "TCPv4"
	v2ProtocolUDPv6        = "UDPv6"
	v2ProtocolUnixStream   = "UNIXStream"
	v2ProtocolUnixDatagram = "UNIXDatagram"
)

var _ ProxyProtocol = &v2{}

type v2 struct {
	protocol string
	src      net.Addr
	dst      net.Addr
}

func (p *v2) Version() string {
	return "v2"
}

func (p *v2) Protocol() string {
	return p.protocol
}

func (p *v2) Source() net.Addr {
	return p.src
}

func (p *v2) Destination() net.Addr {
	return p.dst
}

func newReaderV2(reader io.Reader) Reader {
	r := &v2Reader{
		reader: headerbuf.NewReader(reader),
		done:   false,
	}

	return r
}

type v2Reader struct {
	reader   *headerbuf.Reader
	done     bool
	protocol *v2
	err      error
}

func (rd *v2Reader) ProxyProtocol() (ProxyProtocol, bool) {
	if !rd.done {
		n, err := rd.readProtocol()

		rd.err = err
		_ = rd.reader.Done(n)
	}

	return rd.protocol, rd.protocol != nil
}

func (rd *v2Reader) Error() error {
	return rd.err
}

func (rd *v2Reader) Read(p []byte) (n int, err error) {
	// read header if required
	rd.ProxyProtocol()

	return rd.reader.Read(p)
}

func (rd *v2Reader) readProtocol() (int, error) {
	defer func() {
		rd.done = true
	}()

	// signature
	signature, err := rd.reader.Peek(12)
	if err != nil {
		return 0, errors.New("failed reading signature")
	}

	// signature
	if !bytes.Equal(signature, v2Signature) {
		return 0, errors.New("signature does not match")
	}

	// protocol version and command
	protocolVersionAndCommand, err := rd.reader.PeekByte()
	if err != nil {
		return 0, fmt.Errorf("failed reading protocol version and command: %w", err)
	}

	switch protocolVersionAndCommand {
	case v2ProtocolVersionAndCommandLocal:
		// local
	case v2ProtocolVersionAndCommandProxy:
		// proxy
	default:
		return 0, errors.New("unknown protocol version and command")
	}

	// address family and protocol
	protocol := ""
	unspec := false
	tcp := false
	udp := false
	unix := false
	ipv4 := false
	ipv6 := false
	stream := false
	datagram := false

	var addressFamilyAndProtocol byte
	addressFamilyAndProtocol, err = rd.reader.PeekByte()
	if err != nil {
		return 0, fmt.Errorf("failed reading familiy and protocol: %w", err)
	}

	switch addressFamilyAndProtocol {
	case v2TransportProtocolAndAddressFamilyUNSPEC:
		unspec = true
		protocol = v2ProtocolUNSPEC
	case v2TransportProtocolAndAddressFamilyTCPv4:
		tcp = true
		ipv4 = true
		protocol = v2ProtocolTCPv4
	case v2TransportProtocolAndAddressFamilyUDPv4:
		udp = true
		ipv4 = true
		protocol = v2ProtocolUDPv4
	case v2TransportProtocolAndAddressFamilyTCPv6:
		tcp = true
		ipv6 = true
		protocol = v2ProtocolTCPv6
	case v2TransportProtocolAndAddressFamilyUDPv6:
		udp = true
		ipv6 = true
		protocol = v2ProtocolUDPv6
	case v2TransportProtocolAndAddressFamilyUnixStream:
		unix = true
		stream = true
		protocol = v2ProtocolUnixStream
	case v2TransportProtocolAndAddressFamilyUnixDatagram:
		unix = true
		datagram = true
		protocol = v2ProtocolUnixDatagram
	default:
		return 0, errors.New("unknown address family and protocol")
	}

	// length
	lengthBytes, err := rd.reader.Peek(2)
	if err != nil {
		return 0, fmt.Errorf("failed reading length: %w", err)
	}

	var length uint16
	if err := binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &length); err != nil {
		return 0, fmt.Errorf("failed reading length in network order: %w", err)
	}

	// for TCP/UDP over IPv4, len = 12
	if ipv4 && length < 12 {
		return 0, errors.New("length must be greater than 12 for IPv4")
	}

	// for TCP/UDP over IPv6, len = 36
	if ipv6 && length < 36 {
		return 0, errors.New("length must be greater than 36 for IPv6")
	}

	// for AF_UNIX sockets, len = 216
	if unix && length < 216 {
		return 0, errors.New("length must be greater than 216 for UNIX")
	}

	// for UNSPEC: length anything? is this true? do TLVs make sense in this case
	if unspec && length == 0 {
		// we found a valid proxy protocol header for UNSPEC
		rd.protocol = &v2{
			protocol: protocol,
			src:      nil,
			dst:      nil,
		}

		return -1, nil
	}

	// read all
	buf, err := rd.reader.Peek(int(length))
	if err != nil {
		return 0, fmt.Errorf("failed reading variable payload(%d): %w", int(length), err)
	}

	if len(buf) != int(length) {
		return 0, fmt.Errorf("expected payload siye %d: but got %d", int(length), len(buf))
	}

	// read and ignore TLV data
	if !unspec {
		if ipv4 {
			payload := make([]byte, 12)

			err := binary.Read(bytes.NewReader(buf[0:12]), binary.BigEndian, &payload)
			if err != nil {
				return -1, fmt.Errorf("could not read ipv4 address and ports as big endian: %w", err)
			}

			srcIP := net.IP(payload[0:4])
			dstIP := net.IP(payload[4:8])
			srcPort := binary.BigEndian.Uint16(payload[8:10])
			dstPort := binary.BigEndian.Uint16(payload[10:12])

			if tcp {
				rd.protocol = &v2{
					protocol: protocol,
					src: &net.TCPAddr{
						IP:   srcIP,
						Port: int(srcPort),
					},
					dst: &net.TCPAddr{
						IP:   dstIP,
						Port: int(dstPort),
					},
				}
			}

			if udp {
				rd.protocol = &v2{
					protocol: protocol,
					src: &net.UDPAddr{
						IP:   srcIP,
						Port: int(srcPort),
					},
					dst: &net.UDPAddr{
						IP:   dstIP,
						Port: int(dstPort),
					},
				}
			}

			return -1, nil

		} else if ipv6 {
			payload := make([]byte, 36)

			err := binary.Read(bytes.NewReader(buf[0:36]), binary.BigEndian, &payload)
			if err != nil {
				return -1, fmt.Errorf("could not read ipv6 address and ports as big endian: %w", err)
			}

			srcIP := net.IP(payload[0:16])
			dstIP := net.IP(payload[16:32])
			srcPort := binary.BigEndian.Uint16(payload[32:34])
			dstPort := binary.BigEndian.Uint16(payload[34:36])

			if tcp {
				rd.protocol = &v2{
					protocol: protocol,
					src: &net.TCPAddr{
						IP:   srcIP,
						Port: int(srcPort),
					},
					dst: &net.TCPAddr{
						IP:   dstIP,
						Port: int(dstPort),
					},
				}
			}

			if udp {
				rd.protocol = &v2{
					protocol: protocol,
					src: &net.UDPAddr{
						IP:   srcIP,
						Port: int(srcPort),
					},
					dst: &net.UDPAddr{
						IP:   dstIP,
						Port: int(dstPort),
					},
				}
			}

			return -1, nil
		} else if unix {
			payload := make([]byte, 216)

			err := binary.Read(bytes.NewReader(buf[0:216]), binary.BigEndian, &payload)
			if err != nil {
				return -1, fmt.Errorf("could not read unix address as big endian: %w", err)
			}

			srcBytes := payload[0:108]
			dstBytes := payload[108:216]

			srcEnd := bytes.IndexByte(srcBytes, 0)
			dstEnd := bytes.IndexByte(dstBytes, 0)

			if srcEnd < 0 {
				srcEnd = len(srcBytes)
			}
			if dstEnd < 0 {
				srcEnd = len(dstBytes)
			}

			if stream {
				rd.protocol = &v2{
					protocol: protocol,
					src: &net.UnixAddr{
						Name: string(srcBytes[0:srcEnd]),
						Net:  "unix",
					},
					dst: &net.UnixAddr{
						Name: string(dstBytes[0:dstEnd]),
						Net:  "unix",
					},
				}
			}

			if datagram {
				rd.protocol = &v2{
					protocol: protocol,
					src: &net.UnixAddr{
						Name: string(srcBytes[0:srcEnd]),
						Net:  "unixgram",
					},
					dst: &net.UnixAddr{
						Name: string(dstBytes[0:dstEnd]),
						Net:  "unixgram",
					},
				}
			}

			return -1, nil
		}
	}

	// unspec with tlv? is this possible? should we return an error instead
	return -1, nil
}
