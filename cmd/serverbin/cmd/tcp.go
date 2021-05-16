package cmd

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marsom/serverbin/internal/core"
	"github.com/marsom/serverbin/internal/server"
	"github.com/marsom/serverbin/internal/tcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type TcpCmd struct {
	Address           string `kong:"help='Listen address.',default=':8080'"`
	ManagementAddress string `kong:"help='Readiness, liveness and metric listen address.',default=':8081'"`

	// server
	MaxBufferSize                 int64         `kong:"group='Server',help='Max buffer size in bytes.',default='1024'"`
	ServerTrustedAddresses        []*net.IPNet  `kong:"group='Server',help='Trusted addresses that are known to send correct headers.',default='0.0.0.0/0,::0/0'"`
	ServerShutdownDelay           time.Duration `kong:"group='Server',help='Delay shutdown and let a load balancer remove traffic from this backend.',default='2s'"`
	ServerGracefulShutdownTimeout time.Duration `kong:"group='Server',help='Graceful shutdown time.',default='2m'"`
}

func (cmd *TcpCmd) Run() error {
	if cmd.Address == cmd.ManagementAddress {
		return errors.New("address and management address must be different for a tcp server")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readinessHandler, readinessOn, _ := core.StateHandler()
	livenessHandler, livenessOn, _ := core.StateHandler()
	livenessOn()
	readinessOn()

	srv := &server.TcpServer{
		Name:                    "tcp",
		Address:                 cmd.Address,
		ShutdownDelay:           cmd.ServerShutdownDelay,
		GracefulShutdownTimeout: cmd.ServerGracefulShutdownTimeout,
		RequestHandler: tcp.NewRequestHandler(tcp.Config{
			Server: tcp.Server{
				MaxBufferSize:    cmd.MaxBufferSize,
				TrustedAddresses: cmd.ServerTrustedAddresses,
			},
		}),
	}

	go func() {
		managementMux := http.NewServeMux()
		managementMux.Handle("/-/metrics", promhttp.Handler())
		managementMux.HandleFunc("/-/readiness", readinessHandler)
		managementMux.HandleFunc("/-/liveness", livenessHandler)

		srv := server.HttpServer{
			Name:                    "management",
			Address:                 cmd.ManagementAddress,
			ShutdownDelay:           0 * time.Second,
			GracefulShutdownTimeout: 10 * time.Second,
			Handler:                 managementMux,
		}

		if err := srv.ListenAndServe(ctx); err != nil {
			log.Fatalf("managemnt server failed: %s", err)
		}
	}()

	return srv.ListenAndServe(ctx)
}
