package core

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

func FindBaseUrl(r *http.Request, baseUrl *url.URL, trusted []*net.IPNet) (*url.URL, error) {
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	if ip == nil {
		return baseUrl, errors.New("remote ip not found")
	}

	for _, ipnet := range trusted {
		if ipnet.Contains(ip) {
			proto := "http"
			if s := r.Header.Get("X-Forwarded-Proto"); s != "" {
				proto = s
			}

			host := r.Header.Get("X-Forwarded-Host")

			port := ""
			if s := r.Header.Get("X-Forwarded-Port"); s != "" {
				if proto == "http" && s != "80" {
					port = ":" + s
				}
				if proto == "https" && s != "443" {
					port = ":" + s
				}
			}

			if host != "" {
				return url.Parse(proto + "://" + host + port)
			}
		}
	}

	return baseUrl, errors.New("remote ip not trusted")
}

func StateHandler() (handler func(w http.ResponseWriter, req *http.Request), on, off func()) {
	state := int32(1)

	on = func() {
		atomic.StoreInt32(&state, 0)
	}

	off = func() {
		atomic.StoreInt32(&state, 1)
	}

	handler = func(w http.ResponseWriter, req *http.Request) {
		if atomic.LoadInt32(&state) == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}

	return handler, on, off
}
