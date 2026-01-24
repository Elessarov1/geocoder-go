package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Registrar func(ctx context.Context, s *grpc.Server) error

type StdOptions struct {
	Register Registrar

	// Optional: extra grpc.NewServer options (interceptors, creds, etc).
	ServerOptions []grpc.ServerOption

	// Optional: runtime lifecycle logger.
	Logger func(msg string, kv ...any)
}

type StdGRPCRuntime struct {
	cfg  Config
	opts StdOptions

	srv  *grpc.Server
	lis  net.Listener
	done chan error
}

func NewStdGRPCRuntime(ctx context.Context, cfg Config, opts StdOptions) (*StdGRPCRuntime, error) {
	if opts.Register == nil {
		return nil, fmt.Errorf("grpc.std: Register is required")
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("grpc.std: listen %s: %w", addr, err)
	}

	srv := grpc.NewServer(opts.ServerOptions...)

	// Register service handlers.
	if err := opts.Register(ctx, srv); err != nil {
		_ = lis.Close()
		return nil, err
	}

	// Reflection.
	if cfg.Reflection.Enabled {
		reflection.Register(srv)
	}

	if opts.Logger != nil {
		opts.Logger("gRPC server configured",
			"addr", addr,
			"reflection_enabled", cfg.Reflection.Enabled,
			"shutdown_timeout", cfg.ShutdownTimeout.String(),
		)
	}

	return &StdGRPCRuntime{
		cfg:  cfg,
		opts: opts,
		srv:  srv,
		lis:  lis,
		done: make(chan error, 1),
	}, nil
}

func (r *StdGRPCRuntime) Serve(ctx context.Context) error {
	// Shutdown watcher.
	go func() {
		<-ctx.Done()
		_ = r.shutdown(context.Background())
	}()

	if r.opts.Logger != nil {
		r.opts.Logger("Starting gRPC server", "addr", r.lis.Addr().String())
	}

	err := r.srv.Serve(r.lis)
	r.done <- err
	return err
}

func (r *StdGRPCRuntime) Done() <-chan error {
	return r.done
}

// Graceful shutdown with timeout; fallback to Stop().
func (r *StdGRPCRuntime) Shutdown(ctx context.Context) error {
	return r.shutdown(ctx)
}

func (r *StdGRPCRuntime) shutdown(ctx context.Context) error {
	timeout := r.cfg.ShutdownTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if r.opts.Logger != nil {
		r.opts.Logger("Shutting down gRPC server",
			"addr", r.lis.Addr().String(),
			"timeout", timeout.String(),
		)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		r.srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		if r.opts.Logger != nil {
			r.opts.Logger("gRPC server shutdown complete", "addr", r.lis.Addr().String())
		}
	case <-shutdownCtx.Done():
		if r.opts.Logger != nil {
			r.opts.Logger("gRPC graceful stop timeout, forcing stop", "addr", r.lis.Addr().String())
		}
		r.srv.Stop()
	}

	// Ensure listener is closed (Serve will return).
	if r.lis != nil {
		_ = r.lis.Close()
	}

	return nil
}

// Optional Close() for Stop() duck-typing.
func (r *StdGRPCRuntime) Close() error {
	r.srv.Stop()
	if r.lis != nil {
		return r.lis.Close()
	}
	return nil
}
