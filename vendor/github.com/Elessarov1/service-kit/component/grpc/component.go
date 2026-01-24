package grpc

import (
	"context"

	"github.com/Elessarov1/service-kit/core"
	"github.com/Elessarov1/service-kit/keys"
)

// Runtime is implemented by a gRPC server runtime.
// ListenAndServe must block until server stops.
type Runtime interface {
	Serve(ctx context.Context) error
}

type Component struct {
	cfg       Config
	bootstrap Bootstrap

	rt   Runtime
	done chan error
}

func New(cfg Config, bootstrap Bootstrap) *Component {
	return &Component{
		cfg:       cfg,
		bootstrap: bootstrap,
		done:      make(chan error, 1),
	}
}

func (c *Component) Start(ctx context.Context) error {
	rt, err := c.bootstrap(ctx, c.cfg)
	if err != nil {
		return err
	}
	c.rt = rt

	go func() {
		// Runtime Serve is blocking.
		c.done <- rt.Serve(ctx)
	}()

	return nil
}

func (c *Component) Stop(ctx context.Context) error {
	return core.StopRuntime(ctx, c.rt)
}

func (c *Component) Done() <-chan error { return c.done }
func (c *Component) Name() string       { return keys.GRPC }
