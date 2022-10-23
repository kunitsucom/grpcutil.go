package errorhandler

import (
	"context"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(errorHandler func(ctx context.Context, info *grpc.UnaryServerInfo, err error) error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, errorHandler(ctx, info, err)
		}

		return resp, nil
	}
}

func StreamServerInterceptor(errorHandler func(ctx context.Context, info *grpc.StreamServerInfo, err error) error) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := handler(srv, ss); err != nil {
			return errorHandler(ss.Context(), info, err)
		}

		return nil
	}
}
