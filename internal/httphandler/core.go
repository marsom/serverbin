package httphandler

import (
	"net/http"
	"path"
	"strconv"
)

func RegisterHandlers(serverMux *http.ServeMux, configs ...Config) {
	for _, config := range configs {
		var pattern string

		root := config.Path

		// methods
		pattern = path.Join(root, "method") + "/"
		serverMux.Handle(pattern, &methodHandler{
			Server:        config.Server,
			MethodPattern: defaultMethodPattern,
			Pattern:       pattern,
		})

		// status
		for i := 200; i <= 299; i++ {
			pattern = path.Join(root, "status", strconv.Itoa(i))
			serverMux.HandleFunc(pattern, status(config.Server, i))
		}
		for i := 400; i <= 599; i++ {
			pattern = path.Join(root, "status", strconv.Itoa(i))
			serverMux.HandleFunc(pattern, status(config.Server, i))
		}

		// delay
		if config.Delay != nil {
			pattern = path.Join(root, "delay") + "/"
			serverMux.Handle(pattern, &delayHandler{
				Server:  config.Server,
				Delay:   *config.Delay,
				Pattern: pattern,
			})
		}

		// cookies
		if config.Cookie != nil {
			pattern = path.Join(root, "cookies")
			serverMux.Handle(pattern, &cookieHandler{
				Server: config.Server,
				Cookie: *config.Cookie,
				Path:   root,
			})
		}

		// slow
		if config.Slow != nil {
			pattern = path.Join(root, "slow") + "/"
			serverMux.Handle(pattern, &slowHandler{
				Server:  config.Server,
				Slow:    *config.Slow,
				Pattern: pattern,
			})
		}

		// redirects
		pattern = path.Join(root, "redirect") + "/url/"
		serverMux.Handle(pattern, redirectHandler{
			Server:   config.Server,
			Redirect: *config.Redirect,
			Pattern:  pattern,
			Mode:     redirectByUrl,
		})

		pattern = path.Join(root, "redirect") + "/absolute/"
		serverMux.Handle(pattern, redirectHandler{
			Server:   config.Server,
			Redirect: *config.Redirect,
			Pattern:  pattern,
			Mode:     redirectByAbsolutePath,
		})

		pattern = path.Join(root, "redirect") + "/relative/"
		serverMux.Handle(pattern, redirectHandler{
			Server:   config.Server,
			Redirect: *config.Redirect,
			Pattern:  pattern,
			Mode:     redirectByRelativePath,
		})
	}

}
