package runtimez

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	gw_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type registerHandler = func(ctx context.Context, mux *gw_runtime.ServeMux, conn *grpc.ClientConn) (err error)

func RegisterHandlers(ctx context.Context, mux *gw_runtime.ServeMux, conn *grpc.ClientConn, handlers ...registerHandler) error {
	for _, handler := range handlers {
		if err := handler(ctx, mux, conn); err != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
			return fmt.Errorf("%s: %w", funcName, err)
		}
	}

	return nil
}

type registerHandlerFromEndpoint = func(ctx context.Context, mux *gw_runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

func RegisterHandlersFromEndpoint(ctx context.Context, mux *gw_runtime.ServeMux, endpoint string, opts []grpc.DialOption, handlers ...registerHandlerFromEndpoint) error {
	for _, handler := range handlers {
		if err := handler(ctx, mux, endpoint, opts); err != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
			return fmt.Errorf("%s: %w", funcName, err)
		}
	}

	return nil
}
