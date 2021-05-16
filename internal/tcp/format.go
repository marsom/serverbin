package tcp

import (
	"encoding/json"
	"log"
	"net"
	"strings"
)

func NewRequestHandler(config Config) func(conn net.Conn) {
	return func(conn net.Conn) {
		defer func(conn net.Conn) {
			if err := conn.Close(); err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Printf("close tcp connection failed: %s", err)
				}
			}
		}(conn)

		if resp := newResponse(config, conn); resp != nil {
			jsonWriter := json.NewEncoder(conn)
			jsonWriter.SetIndent("", " ")

			err := jsonWriter.Encode(resp)
			if err != nil {
				log.Printf("could not write to resonse body: %s", err)
			}

			return
		}

		log.Printf("resp was nil")
	}
}
