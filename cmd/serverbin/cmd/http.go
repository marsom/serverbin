package cmd

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/marsom/serverbin/internal/core"
	"github.com/marsom/serverbin/internal/httphandler"
	"github.com/marsom/serverbin/internal/server"
	"github.com/marsom/serverbin/internal/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpCmd struct {
	Address           string   `kong:"help='Listen address.',default=':8080'"`
	ManagementAddress string   `kong:"help='Readiness, liveness and metric listen address.',default=':8081'"`
	Context           []string `kong:"help='Run api on multiple paths.',default='/,/a,/b'"`

	// cookies
	Cookie         bool     `kong:"group='Cookies',help='Enable/Disable cookies.',default='true'"`
	CookieNames    []string `kong:"group='Cookies',help='Cookie names.',default='a,b,c'"`
	CookieHttpOnly bool     `kong:"group='Cookies',help='Set the HttpOnly flag.',default='true'"`

	// delay
	Delay    bool          `kong:"group='Delay',help='Enable/Disable delayed requests.',default='true'"`
	DelayMax time.Duration `kong:"group='Delay',help='Maximum allowed delay.',default='10m'"`

	// slow
	Slow    bool          `kong:"group='Slow',help='Enable/Disable slowed requests .',default='true'"`
	SlowMax time.Duration `kong:"group='Slow',help='Maximum allowed delay.',default='10m'"`

	// redirects
	Redirect    bool `kong:"group='Redirects',help='Enable/Disable redirect requests .',default='true'"`
	RedirectMax uint `kong:"group='Redirects',help='Maximum allowed redirects.',default='20'"`

	// server
	MaxRequestBody                int64         `kong:"group='Server',help='Max request body size in bytes.',default='1048576'"`
	ServerTrustedAddresses        []*net.IPNet  `kong:"group='Server',help='Trusted addresses that are known to send correct headers.',default='0.0.0.0/0,::0/0'"`
	ServerShutdownDelay           time.Duration `kong:"group='Server',help='Delay shutdown and let a load balancer remove traffic from this backend.',default='2s'"`
	ServerGracefulShutdownTimeout time.Duration `kong:"group='Server',help='Graceful shutdown time.',default='2m'"`
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

func findBaseUrl(s string) (*url.URL, error) {
	fields := strings.SplitN(s, ":", 2)
	if fields[0] == "" {
		return url.Parse("http://localhost:" + fields[1])
	}

	//goland:noinspection HttpUrlsUsage
	return url.Parse("http://" + fields[0] + ":" + fields[1])
}

func (r *HttpCmd) Run() error {
	// @TODO: validate context field

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return serve(ctx, r)
}

func serve(ctx context.Context, cmd *HttpCmd) (err error) {
	baseUrl, err := findBaseUrl(cmd.Address)
	if err != nil {
		return err
	}

	managementBaseUrl, err := findBaseUrl(cmd.ManagementAddress)
	if err != nil {
		return err
	}

	readinessHandler, readinessOn, readinessOff := core.StateHandler()
	livenessHandler, livenessOn, _ := core.StateHandler()
	livenessOn()

	var configs []httphandler.Config
	for _, c := range cmd.Context {
		config := httphandler.Config{
			Path: c,
			Server: httphandler.Server{
				MaxRequestBody:    cmd.MaxRequestBody,
				BaseUrl:           baseUrl,
				ManagementBaseUrl: managementBaseUrl,
				TrustedAddresses:  cmd.ServerTrustedAddresses,
			},
		}

		if cmd.Cookie {
			config.Cookie = &httphandler.Cookie{
				Names:      cmd.CookieNames,
				HttpOnly:   cmd.CookieHttpOnly,
				Secure:     false,
				Customizer: nil,
			}
		}

		if cmd.Delay {
			config.Delay = &httphandler.Delay{
				MaxDuration: cmd.DelayMax,
			}
		}

		if cmd.Slow {
			config.Slow = &httphandler.Slow{
				MaxDuration: cmd.SlowMax,
			}
		}

		if cmd.Redirect {
			config.Redirect = &httphandler.Redirect{
				Max: cmd.RedirectMax,
			}
		}

		configs = append(configs, config)
	}

	mux := http.NewServeMux()
	if cmd.Address == cmd.ManagementAddress {
		mux.Handle("/-/metrics", promhttp.Handler())
		mux.HandleFunc("/-/readiness", readinessHandler)
		mux.HandleFunc("/-/liveness", livenessHandler)
	} else {
		go func() {
			managementMux := http.NewServeMux()

			managementMux.Handle("/-/metrics", corsHandler(promhttp.Handler()))
			managementMux.Handle("/-/readiness", corsHandler(http.HandlerFunc(readinessHandler)))
			managementMux.Handle("/-/liveness", corsHandler(http.HandlerFunc(livenessHandler)))

			srv := server.HttpServer{
				Name:                    "management",
				Address:                 cmd.ManagementAddress,
				ShutdownDelay:           0 * time.Second,
				GracefulShutdownTimeout: 3 * time.Second,
				Handler:                 managementMux,
			}

			if err := srv.ListenAndServe(ctx); err != nil {
				log.Fatalf("managemnt server failed: %s", err)
			}
		}()
	}

	mux.Handle("/", swagger.MustUiHandler())
	mux.Handle("/swagger-config.yaml", swagger.ConfigHandler("api.yaml", configs...))
	mux.Handle("/apis.yaml", swagger.DefinitionHandler(configs...))
	mux.Handle("/management-api.yaml", swagger.ManagementDefinitionHandler(configs...))

	for _, config := range configs {
		mux.Handle(path.Join(config.Path, "api.yaml"), swagger.DefinitionHandler(config))
	}

	httphandler.RegisterHandlers(mux, configs...)

	srv := server.HttpServer{
		Name:                    "http",
		Address:                 cmd.Address,
		ShutdownDelay:           cmd.ServerShutdownDelay,
		GracefulShutdownTimeout: cmd.ServerGracefulShutdownTimeout,
		Handler:                 mux,
		ReadinessOn:             readinessOn,
		ReadinessOff:            readinessOff,
	}

	return srv.ListenAndServe(ctx)
}
