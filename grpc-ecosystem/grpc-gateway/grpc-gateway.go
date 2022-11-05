package grpcgateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	signalz "github.com/kunitsuinc/util.go/os/signal"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"

	grpcz "github.com/kunitsuinc/grpcutil.go/grpc"
)

type server struct {
	httpServer            *http.Server
	signalChan            chan os.Signal
	continueSignalHandler func(sig os.Signal) bool
	shutdownTimeout       time.Duration
	shutdownErrorHandler  func(err error)
}

type ServeOption func(s *server)

func WithSignalChannel(signalChan chan os.Signal) ServeOption {
	return func(s *server) { s.signalChan = signalChan }
}

func WithContinueSignalHandler(continueSignalHandler func(sig os.Signal) bool) ServeOption {
	return func(s *server) { s.continueSignalHandler = continueSignalHandler }
}

func WithShutdownErrorHandler(shutdownErrorHandler func(err error)) ServeOption {
	return func(s *server) { s.shutdownErrorHandler = shutdownErrorHandler }
}

// ServeGRPC serve gRPC Server with gRPC Gateway
//
//nolint:funlen,cyclop
func ServeGRPC(
	ctx context.Context,
	l net.Listener,
	grpcServer *grpc.Server,
	grpcGatewayMux *http.ServeMux,
	opts ...ServeOption,
) error {
	s := &server{
		httpServer:            &http.Server{ReadHeaderTimeout: 10 * time.Second},
		signalChan:            signalz.Notify(make(chan os.Signal, 1), syscall.SIGHUP, os.Interrupt, syscall.SIGTERM),
		continueSignalHandler: func(sig os.Signal) bool { return sig == syscall.SIGHUP },
		shutdownTimeout:       10 * time.Second,
		shutdownErrorHandler:  func(err error) { log.Println("shutdown:", err) },
	}

	for _, opt := range opts {
		opt(s)
	}

	s.httpServer.Handler = grpcz.GRPCHandler(grpcServer, grpcGatewayMux, &http2.Server{})

	serve := func(errChan chan<- error) {
		errChan <- s.httpServer.Serve(l)
	}

	serveErrChan := make(chan error, 1)
	go serve(serveErrChan)

	shutdown := func() error {
		grpcServer.GracefulStop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server.Shutdown: %w", err)
		}

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			if err := shutdown(); err != nil {
				s.shutdownErrorHandler(err)
			}
		case sig := <-s.signalChan:
			if s.continueSignalHandler(sig) {
				continue
			}
			if err := shutdown(); err != nil {
				s.shutdownErrorHandler(err)
			}
		case err := <-serveErrChan:
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}
			return nil
		}
	}
}
