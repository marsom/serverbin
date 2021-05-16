package httphandler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Delay configuration
type Delay struct {
	MaxDuration time.Duration
}

var _ http.Handler = (*delayHandler)(nil)

type delayHandler struct {
	Server
	Delay
	Pattern string
}

func (d delayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	duration, err := time.ParseDuration(strings.TrimPrefix(r.URL.Path, d.Pattern))
	if err != nil {
		fn := format(d.Server, r, http.StatusBadRequest, errors.New("could not input to a valid duration"), err)
		fn(w, r)

		return
	}

	if duration >= d.MaxDuration {
		fn := format(d.Server, r, http.StatusBadRequest, fmt.Errorf("duration must be less then %s", d.MaxDuration))
		fn(w, r)

		return
	}

	time.Sleep(duration)

	fn := format(d.Server, r, http.StatusOK, nil)
	fn(w, r)
}
