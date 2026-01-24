package config

import (
	"fmt"
	"strconv"
	"strings"
)

func RequireString(m map[string]any, key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("missing key: %s", key)
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("key %s must be non-empty string", key)
	}
	return strings.TrimSpace(s), nil
}

// RequireInt принимает int/int64/float64 или строку "8080"
func RequireInt(m map[string]any, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing key: %s", key)
	}
	switch t := v.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return 0, fmt.Errorf("key %s must be int, got %q", key, t)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("key %s must be int", key)
	}
}

func OptionalBool(m map[string]any, key string, def bool) bool {
	v, ok := m[key]
	if !ok {
		return def
	}

	if b, ok := v.(bool); ok {
		return b
	}

	if s, ok := v.(string); ok {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true", "1", "yes", "y", "on":
			return true
		case "false", "0", "no", "n", "off":
			return false
		default:
			return def
		}
	}

	return def
}

func OptionalStringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}

	// YAML array => []any
	if arr, ok := v.([]any); ok {
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			if s, ok := x.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
		}
		return out
	}

	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	return nil
}
