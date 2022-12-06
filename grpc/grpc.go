package grpcz

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func ServerWithGatewayHandler(grpcServer *grpc.Server, grpcGatewayHandler http.Handler, http2Server *http2.Server) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: cf. https://github.com/grpc/grpc-go/issues/555#issuecomment-443293451
		// NOTE: cf. https://github.com/philips/grpc-gateway-example/issues/22#issuecomment-490733965
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			grpcGatewayHandler.ServeHTTP(w, r)
		}
	}), http2Server)
}

type ServerWithGateway struct {
	httpServer     *http.Server
	grpcServer     *grpc.Server
	grpcGatewayMux *http.ServeMux

	signalChannel         chan os.Signal
	continueSignalHandler func(sig os.Signal) bool
	shutdownTimeout       time.Duration
	shutdownErrorHandler  func(err error)
}

type ServerWithGatewayOption func(s *ServerWithGateway)

// WithHTTPServer
//
// If *http.Server has (*http.Server).Handler, it is ignored by grpcz.ServerWithGateway.
func WithHTTPServer(httpServer *http.Server) ServerWithGatewayOption {
	return func(s *ServerWithGateway) { s.httpServer = httpServer }
}

func WithSignalChannel(signalChannel chan os.Signal) ServerWithGatewayOption {
	return func(s *ServerWithGateway) { s.signalChannel = signalChannel }
}

func WithContinueSignalHandler(continueSignalHandler func(sig os.Signal) bool) ServerWithGatewayOption {
	return func(s *ServerWithGateway) { s.continueSignalHandler = continueSignalHandler }
}

func WithShutdownTimeout(shutdownTimeout time.Duration) ServerWithGatewayOption {
	return func(s *ServerWithGateway) { s.shutdownTimeout = shutdownTimeout }
}

func WithShutdownErrorHandler(shutdownErrorHandler func(err error)) ServerWithGatewayOption {
	return func(s *ServerWithGateway) { s.shutdownErrorHandler = shutdownErrorHandler }
}

func NewServerWithGateway(grpcServer *grpc.Server, grpcGatewayMux *http.ServeMux, opts ...ServerWithGatewayOption) *ServerWithGateway {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)

	s := &ServerWithGateway{
		httpServer:     &http.Server{ReadHeaderTimeout: 10 * time.Second},
		grpcServer:     grpcServer,
		grpcGatewayMux: grpcGatewayMux,

		signalChannel:         signalChannel,
		continueSignalHandler: func(sig os.Signal) bool { return sig == syscall.SIGHUP },
		shutdownTimeout:       10 * time.Second,
		shutdownErrorHandler:  func(err error) { log.Println("shutdown:", err) },
	}

	for _, opt := range opts {
		opt(s)
	}

	s.httpServer.Handler = ServerWithGatewayHandler(s.grpcServer, s.grpcGatewayMux, &http2.Server{})

	return s
}

// ListenAndServe serve gRPC Server with gRPC Gateway.
func (s *ServerWithGateway) Serve(
	ctx context.Context,
	l net.Listener,
) error {
	serve := func(errChan chan<- error) { errChan <- s.httpServer.Serve(l) }

	shutdown := func() error {
		s.grpcServer.GracefulStop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server.Shutdown: %w", err)
		}

		return nil
	}

	serveErrChan := make(chan error, 1)
	go serve(serveErrChan)

	for {
		select {
		case <-ctx.Done():
			if err := shutdown(); err != nil {
				s.shutdownErrorHandler(err)
			}
		case sig := <-s.signalChannel:
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
