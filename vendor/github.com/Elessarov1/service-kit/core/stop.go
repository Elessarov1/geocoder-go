package core

import "context"

func StopReverse(ctx context.Context, comps []Component) {
	for i := len(comps) - 1; i >= 0; i-- {
		_ = comps[i].Stop(ctx)
	}
}
