package grpc_server

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func unaryLoggingInterceptor(lg *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		d := time.Since(start)

		if err != nil {
			lg.Warn("grpc request failed", zap.String("method", info.FullMethod), zap.Duration("dur", d), zap.Error(err))
		} else {
			lg.Info("grpc request", zap.String("method", info.FullMethod), zap.Duration("dur", d))
		}
		return resp, err
	}
}
