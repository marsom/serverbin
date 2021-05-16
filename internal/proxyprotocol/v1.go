package proxyprotocol

import (
	"errors"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/marsom/serverbin/internal/headerbuf"
)

var _ ProxyProtocol = v1{}

type v1 struct {
	protocol string
	src      net.Addr
	dst      net.Addr
}

func (p v1) Version() string {
	return "v1"
}

func (p v1) Protocol() string {
	return p.protocol
}

func (p v1) Source() net.Addr {
	return p.src
}

func (p v1) Destination() net.Addr {
	return p.dst
}

func newReaderV1(reader io.Reader) Reader {
	r := &v1Reader{
		reader: headerbuf.NewReader(reader),
		done:   false,
	}

	return r
}

type v1Reader struct {
	reader   *headerbuf.Reader
	done     bool
	protocol *v1
	err      error
}

func (rd *v1Reader) ProxyProtocol() (ProxyProtocol, bool) {
	if !rd.done {
		n, err := rd.readProtocol()

		rd.err = err
		_ = rd.reader.Done(n)
	}

	return rd.protocol, rd.protocol != nil
}

func (rd *v1Reader) Error() error {
	return rd.err
}

func (rd *v1Reader) Read(p []byte) (n int, err error) {
	// read header if required
	rd.ProxyProtocol()

	return rd.reader.Read(p)
}

func (rd *v1Reader) readProtocol() (int, error) {
	defer func() {
		rd.done = true
	}()

	buffer, err := rd.reader.Peek(107)
	if err != nil {
		return 0, errors.New("failed reading signature")
	}

	if len(buffer) <= 0 {
		return 0, errors.New("failed reading signature")
	}

	line := string(buffer)
	end := strings.Index(line, "\n")

	if end <= 0 {
		return 0, errors.New("missing LF")
	}

	if line[end] == '\r' {
		return 0, errors.New("missing CRLF")
	}

	tokens := strings.Split(strings.TrimSpace(strings.Split(line, "\n")[0]), " ")

	if len(tokens) < 2 {
		return 0, errors.New("expected at least 2 tokens")
	}

	if tokens[0] != "PROXY" {
		return 0, errors.New("signature does not match")
	}

	protocol := tokens[1]
	switch protocol {
	case "TCP4":
	case "TCP6":
	case "UNKNOWN":
		rd.protocol = &v1{
			protocol: "UNKNOWN",
		}

		return end + 1, nil
	default:
		return 0, errors.New("unknown protocol, expected TCP4,TCP6 or UNKNOWN")
	}

	if len(tokens) < 6 {
		return 0, errors.New("expected at most 5 tokens")
	}

	srcIP := net.ParseIP(tokens[2])
	if srcIP == nil {
		return 0, errors.New("could not parse source ip")
	}

	dstIP := net.ParseIP(tokens[3])
	if dstIP == nil {
		return 0, errors.New("could not parse destination ip")
	}

	srcPort, err := strconv.Atoi(tokens[4])
	if err != nil || srcPort < 0 || srcPort > 65535 {
		return 0, errors.New("could not parse source port")
	}

	dstPort, err := strconv.Atoi(tokens[5])
	if err != nil || dstPort < 0 || dstPort > 65535 {
		return 0, errors.New("could not parse destination port")
	}

	rd.protocol = &v1{
		src: &net.TCPAddr{
			IP:   srcIP,
			Port: srcPort,
		},
		dst: &net.TCPAddr{
			IP:   dstIP,
			Port: dstPort,
		},
	}

	return end + 1, nil
}
