package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type TcpServer struct {
	Name                    string
	Address                 string
	ShutdownDelay           time.Duration
	GracefulShutdownTimeout time.Duration
	ReadinessOn             func()
	ReadinessOff            func()
	RequestHandler          func(conn net.Conn)
}

func (s *TcpServer) ListenAndServe(ctx context.Context) error {
	tcpServer := &tcpServer{
		quit:           make(chan interface{}),
		requestHandler: s.RequestHandler,
	}

	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	tcpServer.listener = l
	tcpServer.wg.Add(1)
	go tcpServer.serve()

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

	// exit loop in tcpServer.serve()
	close(tcpServer.quit)
	err = tcpServer.listener.Close()
	if err != nil {
		return fmt.Errorf("%s server stop failed: %w", s.Name, err)
	}

	time.AfterFunc(s.GracefulShutdownTimeout, func() {
		log.Fatalf("%s server shutdown failed (timeout=%s)", s.Name, s.GracefulShutdownTimeout)
	})

	// wait for active connection to finish
	tcpServer.wg.Wait()

	log.Printf("%s server stopped", s.Name)

	return nil
}

type tcpServer struct {
	listener       net.Listener
	quit           chan interface{}
	wg             sync.WaitGroup
	requestHandler func(conn net.Conn)
}

func (s *tcpServer) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Println("accept error (default)", err)
			}
		} else {
			s.wg.Add(1)
			go func() {
				s.requestHandler(conn)
				s.wg.Done()
			}()
		}
	}
}
