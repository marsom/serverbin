package httphandler

import (
	"net/http"
)

func status(config Server, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := format(config, r, statusCode, nil)
		fn(w, r)
	}
}
