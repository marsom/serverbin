package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

type HttpServer struct {
	Name                    string
	Address                 string
	ShutdownDelay           time.Duration
	GracefulShutdownTimeout time.Duration
	ReadinessOn             func()
	ReadinessOff            func()
	Handler                 *http.ServeMux
}

func (s *HttpServer) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.Address,
		Handler: s.Handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("%s server listen failed: %s\n", s.Name, err)
		}
	}()

	// let ListenAndServe fail before we print we are ready
	// not the best way to to this, but it should work in most cases
	time.Sleep(5 * time.Millisecond)

	log.Printf("%s server started on %s", s.Name, s.Address)
	if s.ReadinessOn != nil {
		s.ReadinessOn()
	}

	// block
	<-ctx.Done()

	log.Printf("%s server shutdown initalized (delay=%s)", s.Name, s.ShutdownDelay)
	if s.ReadinessOff != nil {
		s.ReadinessOff()
	}

	time.Sleep(s.ShutdownDelay)

	ctxShutDown, cancel := context.WithTimeout(context.Background(), s.GracefulShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("%s server shutdown failed (timeout=%s): %s", s.Name, s.GracefulShutdownTimeout, err)
	}

	log.Printf("%s server stopped", s.Name)

	return nil
}
