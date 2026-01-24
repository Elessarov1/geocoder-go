package core

import (
	"fmt"
)

// SortDescriptors: topo sort by DependsOn.
func SortDescriptors(descs []*Descriptor) ([]*Descriptor, error) {
	byName := make(map[string]*Descriptor, len(descs))
	inDeg := make(map[string]int, len(descs))
	adj := make(map[string][]string, len(descs))

	for _, d := range descs {
		n := d.Comp.Name()
		if n == "" {
			return nil, fmt.Errorf("component name cannot be empty")
		}
		if _, exists := byName[n]; exists {
			return nil, fmt.Errorf("duplicate component name: %s", n)
		}
		byName[n] = d
		inDeg[n] = 0
	}

	// validate deps exist + build graph
	for _, d := range descs {
		n := d.Comp.Name()
		for _, dep := range d.DependsOn {
			if _, ok := byName[dep]; !ok {
				return nil, fmt.Errorf("%s depends_on unknown component: %s", n, dep)
			}
			adj[dep] = append(adj[dep], n)
			inDeg[n]++
		}
	}

	zero := make([]string, 0)
	for n, deg := range inDeg {
		if deg == 0 {
			zero = append(zero, n)
		}
	}

	// pick minimal by Index for stability
	pick := func(names []string) int {
		best := 0
		for i := 1; i < len(names); i++ {
			if byName[names[i]].Index < byName[names[best]].Index {
				best = i
			}
		}
		return best
	}

	out := make([]*Descriptor, 0, len(descs))
	for len(zero) > 0 {
		i := pick(zero)
		n := zero[i]
		zero = append(zero[:i], zero[i+1:]...)

		out = append(out, byName[n])

		for _, to := range adj[n] {
			inDeg[to]--
			if inDeg[to] == 0 {
				zero = append(zero, to)
			}
		}
	}

	if len(out) != len(descs) {
		return nil, fmt.Errorf("depends_on cycle detected")
	}
	return out, nil
}
