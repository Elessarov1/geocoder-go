package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Elessarov1/service-kit/config"
)

type RunOption func(*runOptions)

type runOptions struct {
	stopTimeout time.Duration
}

// WithStopTimeout sets a timeout for stopping components.
// If 0, StopReverse will be called with context.Background().
func WithStopTimeout(d time.Duration) RunOption {
	return func(o *runOptions) { o.stopTimeout = d }
}

// Run:
//   - reads YAML config (with env expansion)
//   - builds components via registry (validation + topo sort)
//   - starts all components (Start must not block)
//   - waits for ctx.Done OR first component Done()
//   - stops started components in reverse order
//
// Return policy:
//   - ctx cancellation => nil
//   - component Done() returning nil or context.Canceled => nil
//   - otherwise returns the error
func Run(ctx context.Context, configPath string, reg Registry, opts ...RunOption) error {
	var o runOptions
	for _, fn := range opts {
		fn(&o)
	}

	raw, err := config.ReadYAML(configPath)
	if err != nil {
		return err
	}

	comps, err := BuildComponents(raw, reg)
	if err != nil {
		return err
	}
	if len(comps) == 0 {
		return fmt.Errorf("no components configured")
	}

	started := make([]Component, 0, len(comps))
	stopAll := func() {
		if len(started) == 0 {
			return
		}
		if o.stopTimeout <= 0 {
			StopReverse(context.Background(), started)
			return
		}
		stopCtx, cancel := context.WithTimeout(context.Background(), o.stopTimeout)
		defer cancel()
		StopReverse(stopCtx, started)
	}

	// start in order
	for _, c := range comps {
		if err := c.Start(ctx); err != nil {
			stopAll()
			return err
		}
		started = append(started, c)
	}

	// fan-in: first Done() wins
	errCh := make(chan error, 1)
	for _, c := range started {
		cc := c
		go func() {
			e := <-cc.Done()
			select {
			case errCh <- e:
			default:
			}
		}()
	}

	var result error
	select {
	case <-ctx.Done():
		result = nil
	case e := <-errCh:
		if e == nil || errors.Is(e, context.Canceled) {
			result = nil
		} else {
			result = e
		}
	}

	stopAll()
	return result
}
