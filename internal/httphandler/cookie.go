package httphandler

import (
	"fmt"
	"net/http"
	"time"
)

// Cookie configuration
type Cookie struct {
	Names      []string
	HttpOnly   bool
	Secure     bool
	Customizer func(cookie *http.Cookie)
}

type cookieHandler struct {
	Server
	Cookie
	Path string
}

func (c cookieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	customizer := func(cookie *http.Cookie) *http.Cookie {
		if c.Customizer != nil {
			// preserve name and path
			name := cookie.Name
			path := cookie.Path

			c.Customizer(cookie)

			cookie.Name = name
			cookie.Path = path
		}

		return cookie
	}

	switch r.Method {
	case "GET", "HEAD":
		fn := format(c.Server, r, http.StatusOK, nil)
		fn(w, r)
	case "DELETE":
		for _, name := range c.Names {
			http.SetCookie(w, &http.Cookie{
				Name:    name,
				Value:   "",
				Path:    c.Path,
				Expires: time.Unix(0, 0),
			})
		}

		fn := format(c.Server, r, http.StatusOK, nil)
		fn(w, r)
	case "PUT":
		for _, name := range c.Names {
			http.SetCookie(w, customizer(&http.Cookie{
				Name:     name,
				Value:    fmt.Sprintf("test %s", c.Path),
				Path:     c.Path,
				HttpOnly: c.HttpOnly,
				Secure:   c.Secure,
				SameSite: http.SameSiteStrictMode,
			}))
		}

		fn := format(c.Server, r, http.StatusOK, nil)
		fn(w, r)
	default:
		fn := format(c.Server, r, http.StatusMethodNotAllowed, nil)
		fn(w, r)
	}
}

var _ http.Handler = (*cookieHandler)(nil)
