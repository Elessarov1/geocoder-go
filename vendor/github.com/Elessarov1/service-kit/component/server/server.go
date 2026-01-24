package server

import (
	"context"

	"github.com/Elessarov1/service-kit/keys"
)

type Bootstrap func(ctx context.Context, cfg Config) (Runtime, error)

type Server struct {
	cfg       Config
	bootstrap Bootstrap

	rt   Runtime
	done chan error
}

func New(cfg Config, bootstrap Bootstrap) *Server {
	return &Server{
		cfg:       cfg,
		bootstrap: bootstrap,
		done:      make(chan error, 1),
	}
}

func (s *Server) Name() string { return keys.Server }

func (s *Server) Start(ctx context.Context) error {
	rt, err := s.bootstrap(ctx, s.cfg)
	if err != nil {
		return err
	}
	s.rt = rt

	go func() {
		s.done <- rt.ListenAndServe(ctx)
	}()

	return nil
}

// Optional stop interfaces
type shutdownErr interface{ Shutdown(context.Context) error }
type shutdownVoid interface{ Shutdown(context.Context) }
type closeErr interface{ Close() error }
type closeVoid interface{ Close() }

func (s *Server) Stop(ctx context.Context) error {
	if s.rt == nil {
		return nil
	}

	// Prefer graceful shutdown if available
	switch rt := any(s.rt).(type) {
	case shutdownErr:
		return rt.Shutdown(ctx)
	case shutdownVoid:
		rt.Shutdown(ctx)
		return nil
	case closeErr:
		return rt.Close()
	case closeVoid:
		rt.Close()
		return nil
	default:
		// Not stoppable – nothing to do
		return nil
	}
}

func (s *Server) Done() <-chan error { return s.done }
