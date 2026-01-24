package core

import "context"

func StopReverse(ctx context.Context, comps []Component) {
	for i := len(comps) - 1; i >= 0; i-- {
		_ = comps[i].Stop(ctx)
	}
}

// StopRuntime calls one of the optional shutdown methods on runtime in priority order:
//
//  1. Shutdown(context.Context) error
//  2. Shutdown(context.Context)
//  3. Close() error
//  4. Close()
//
// If none are present, it does nothing and returns nil.
func StopRuntime(ctx context.Context, rt any) error {
	if rt == nil {
		return nil
	}

	// 1) Shutdown(ctx) error
	if s, ok := rt.(interface{ Shutdown(context.Context) error }); ok {
		return s.Shutdown(ctx)
	}

	// 2) Shutdown(ctx)
	if s, ok := rt.(interface{ Shutdown(context.Context) }); ok {
		s.Shutdown(ctx)
		return nil
	}

	// 3) Close() error
	if c, ok := rt.(interface{ Close() error }); ok {
		return c.Close()
	}

	// 4) Close()
	if c, ok := rt.(interface{ Close() }); ok {
		c.Close()
		return nil
	}

	return nil
}
