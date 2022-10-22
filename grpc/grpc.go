package grpcz

import (
	"net/http"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func GRPCHandler(grpcServer *grpc.Server, handler http.Handler, http2Server *http2.Server) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: cf. https://github.com/grpc/grpc-go/issues/555#issuecomment-443293451
		// NOTE: cf. https://github.com/philips/grpc-gateway-example/issues/22#issuecomment-490733965
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	}), http2Server)
}
