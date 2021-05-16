package httphandler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type redirectMode int

const (
	redirectByUrl redirectMode = iota
	redirectByAbsolutePath
	redirectByRelativePath
)

// Redirect configuration for redirects
type Redirect struct {
	Max uint
}

var _ http.Handler = (*redirectHandler)(nil)

type redirectHandler struct {
	Server
	Redirect
	Pattern string
	Mode    redirectMode
}

func (h redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	depth, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, h.Pattern), "/"))
	if err != nil {
		fn := format(h.Server, r, http.StatusBadRequest, fmt.Errorf("depth is not an integer: %w", err))
		fn(w, r)

		return
	}

	if depth >= 20 || depth < 0 {
		fn := format(h.Server, r, http.StatusBadRequest, errors.New("max redirect count is 20"))
		fn(w, r)

		return
	}

	if depth == 0 {
		fn := format(h.Server, r, http.StatusOK, nil)
		fn(w, r)

		return
	}

	code := http.StatusFound
	if codeString := r.URL.Query().Get("code"); codeString != "" {
		i, err := strconv.Atoi(codeString)
		if err != nil {
			fn := format(h.Server, r, http.StatusBadRequest, errors.New("invalid redirect code given"))
			fn(w, r)

			return
		}

		if i < 300 || i >= 400 {
			fn := format(h.Server, r, http.StatusBadRequest, errors.New("invalid redirect code given: "+strconv.Itoa(i)))
			fn(w, r)

			return
		}

		code = i
	}

	redirectUrl := url.URL{}

	// TODO: read body / discard body

	switch h.Mode {
	case redirectByUrl:
		redirectUrl.Scheme = h.BaseUrl.Scheme
		redirectUrl.Host = h.BaseUrl.Host
		redirectUrl.Path = path.Join(h.Pattern, strconv.Itoa(depth-1))
		q := redirectUrl.Query()
		q.Set("code", strconv.Itoa(code))
		redirectUrl.RawQuery = q.Encode()

		http.Redirect(w, r, redirectUrl.String(), code)
	case redirectByAbsolutePath:
		redirectUrl.Path = path.Join(h.Pattern, strconv.Itoa(depth-1))
		q := redirectUrl.Query()
		q.Set("code", strconv.Itoa(code))
		redirectUrl.RawQuery = q.Encode()

		http.Redirect(w, r, redirectUrl.String(), code)
	case redirectByRelativePath:
		// adapted from net/http/server.go because it does not support relative paths
		redirectURL := strconv.Itoa(depth - 1)
		redirectURL = redirectURL + "?code=" + strconv.Itoa(code)

		h := w.Header()

		// RFC 7231 notes that a short HTML body is usually included in
		// the response because older user agents may not understand 301/307.
		// Do it only if the request didn't already have a Content-Type header.
		_, hadCT := h["Content-Type"]

		h.Set("Location", redirectURL)
		if !hadCT && (r.Method == "GET" || r.Method == "HEAD") {
			h.Set("Content-Type", "text/html; charset=utf-8")
		}
		w.WriteHeader(code)

		// Shouldn't send the body for POST or HEAD; that leaves GET.
		if !hadCT && r.Method == "GET" {
			body := "<a href=\"" + redirectURL + "\">" + strconv.Itoa(code) + "</a>.\n"
			_, _ = fmt.Fprintln(w, body)
		}
	default:
		http.Redirect(w, r, path.Join("http://localhost:8080", h.Pattern, strconv.Itoa(depth-1)), code)
	}
}
