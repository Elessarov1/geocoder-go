package config

import (
	"fmt"
	"os"
	"regexp"
)

var envRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::([^}]*))?\}`)

// ExpandEnv рекурсивно проходит map/slice и заменяет ${ENV[:default]} в строках.
func ExpandEnv(v any) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			x, err := ExpandEnv(vv)
			if err != nil {
				return nil, err
			}
			out[k] = x
		}
		return out, nil

	case []any:
		out := make([]any, 0, len(t))
		for _, vv := range t {
			x, err := ExpandEnv(vv)
			if err != nil {
				return nil, err
			}
			out = append(out, x)
		}
		return out, nil

	case string:
		return expandString(t)

	default:
		return v, nil
	}
}

func expandString(s string) (string, error) {
	var firstErr error

	res := envRe.ReplaceAllStringFunc(s, func(m string) string {
		if firstErr != nil {
			return m
		}
		sub := envRe.FindStringSubmatch(m)
		env := sub[1]
		def := sub[2]

		val, ok := os.LookupEnv(env)
		if ok {
			return val
		}
		if def != "" {
			return def
		}
		firstErr = fmt.Errorf("missing required env var: %s", env)
		return m
	})

	if firstErr != nil {
		return "", firstErr
	}
	return res, nil
}
