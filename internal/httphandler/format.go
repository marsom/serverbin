package httphandler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

func slowFormat(config Server, d time.Duration, r *http.Request, statusCode int, errs ...error) http.HandlerFunc {
	w := httptest.NewRecorder()

	format(config, r, statusCode, errs...)(w, r)

	resp := w.Result()

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fn := format(config, r, http.StatusInternalServerError, errors.New("could not buffer response"), err)
			fn(w, r)

			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			fn := format(config, r, http.StatusInternalServerError, errors.New("could not buffer response"), err)
			fn(w, r)

			return
		}

		sleep := int(d) / len(body)

		h := w.Header()
		for key, values := range resp.Header {
			for _, value := range values {
				h.Set(key, value)
			}
		}

		w.WriteHeader(statusCode)
		flusher.Flush()

		for i := range body {
			_, _ = w.Write([]byte{body[i]})

			time.Sleep(time.Duration(sleep))

			flusher.Flush()
		}

	}
}

func format(config Server, r *http.Request, statusCode int, errs ...error) http.HandlerFunc {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept

	r.Body = http.MaxBytesReader(nil, r.Body, config.MaxRequestBody)

	accept := r.Header.Get("Accept")

	resp := newResponse(config, r, errs...)

	// return json if nothing is specified
	if accept == "" {
		return formatJSON(statusCode, resp)
	}

	// this is a poor man's negotiation version, but for the start it does it's job
	// https://github.com/golang/go/issues/19307
	// https://pkg.go.dev/golang.org/x/text/language#ParseAcceptLanguage
	for _, t := range strings.Split(accept, ",") {
		switch strings.SplitN(strings.TrimSpace(t), ";", 2)[0] {
		case "*/*":
			return formatJSON(statusCode, resp)
		case "application/json":
			return formatJSON(statusCode, resp)
		case "text/plain":
			return formatTEXT(statusCode, resp)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// json
func formatJSON(statusCode int, resp *response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(statusCode)
		jsonWriter := json.NewEncoder(w)
		jsonWriter.SetIndent("", " ")
		err := jsonWriter.Encode(resp)

		if err != nil {
			log.Printf("could not write to resonse body: %s", err)
		}
	}
}

// text
func formatTEXT(statusCode int, resp *response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)

		if len(resp.Errors) > 0 {
			_, _ = w.Write([]byte("# Errors\n\n"))
			for _, err := range resp.Errors {
				_, _ = w.Write([]byte("- "))
				_, _ = w.Write([]byte(err))
				_, _ = w.Write([]byte("\n"))
			}
			_, _ = w.Write([]byte("\n\n"))
		}

		if len(resp.Headers) > 0 {
			_, _ = w.Write([]byte("# Headers\n\n"))
			for header, values := range resp.Headers {
				_, _ = w.Write([]byte(header))
				_, _ = w.Write([]byte(":\n"))

				for _, value := range values {
					_, _ = w.Write([]byte("- "))
					_, _ = w.Write([]byte(value))
					_, _ = w.Write([]byte("\n"))
				}
			}
			_, _ = w.Write([]byte("\n\n"))
		}

		if len(resp.Cookies) > 0 {
			_, _ = w.Write([]byte("# Cookies\n\n"))
			for _, cookie := range resp.Cookies {
				_, _ = w.Write([]byte("- name: "))
				_, _ = w.Write([]byte(cookie.Name))
				_, _ = w.Write([]byte("\n"))
				_, _ = w.Write([]byte("  value: "))
				_, _ = w.Write([]byte(cookie.Value))
				_, _ = w.Write([]byte("\n"))

			}
			_, _ = w.Write([]byte("\n\n"))
		}
	}
}
