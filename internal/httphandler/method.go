package httphandler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	defaultMethodPattern = regexp.MustCompile("[a-zA-Z]{3,7}")
)

var _ http.Handler = (*methodHandler)(nil)

type methodHandler struct {
	Server
	MethodPattern *regexp.Regexp
	Pattern       string
}

func (h methodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(strings.TrimPrefix(r.URL.Path, h.Pattern))

	if method != r.Method {
		fn := format(h.Server, r, http.StatusMethodNotAllowed, fmt.Errorf("only %s requests are allowed", method))
		fn(w, r)

		return
	}

	if !h.MethodPattern.MatchString(method) {
		fn := format(h.Server, r, http.StatusBadRequest,
			errors.New("given method must only contain characters from A-Z"),
			errors.New("must be at least 3 but not more than 7 characters long"),
		)
		fn(w, r)

		return
	}

	// return data
	fn := format(h.Server, r, http.StatusOK, nil)
	fn(w, r)
}
