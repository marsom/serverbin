package httphandler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Slow configuration
type Slow struct {
	MaxDuration time.Duration
}

var _ http.Handler = (*slowHandler)(nil)

type slowHandler struct {
	Server
	Slow
	Pattern string
}

func (h slowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	duration, err := time.ParseDuration(strings.TrimPrefix(r.URL.Path, h.Pattern))
	if err != nil {
		fn := format(h.Server, r, http.StatusBadRequest, errors.New("duration is invalid"), err)
		fn(w, r)

		return
	}

	if duration >= h.MaxDuration {
		fn := format(h.Server, r, http.StatusBadRequest, fmt.Errorf("duration must be less then %s", h.MaxDuration))
		fn(w, r)

		return
	}

	fn := slowFormat(h.Server, duration, r, http.StatusOK)
	fn(w, r)
}
