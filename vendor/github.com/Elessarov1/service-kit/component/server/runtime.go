package server

import "context"

// Runtime requires only the start method.
// Shutdown/Close are optional and will be detected dynamically in Stop().
type Runtime interface {
	ListenAndServe(ctx context.Context) error
}
