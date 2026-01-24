package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

func OptionalSection(sec map[string]any, key string) (map[string]any, bool, error) {
	v, ok := sec[key]
	if !ok || v == nil {
		return nil, false, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("key %s must be map/object", key)
	}
	return m, true, nil
}

func RequireSection(sec map[string]any, key string) (map[string]any, error) {
	m, ok, err := OptionalSection(sec, key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("missing key: %s", key)
	}
	return m, nil
}

func OptionalString(sec map[string]any, key string, def string) string {
	v, ok := sec[key]
	if !ok || v == nil {
		return def
	}
	s, ok := v.(string)
	if !ok {
		return def
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}

func RequireDurationString(sec map[string]any, key string) (string, error) {
	v, ok := sec[key]
	if !ok || v == nil {
		return "", fmt.Errorf("missing key: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("key %s must be duration string", key)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("key %s must be non-empty duration string", key)
	}
	return s, nil
}

// OptionalMap returns a nested map section if present.
// Supports both map[string]any and map[any]any (in case of YAML decoder differences).
func OptionalMap(sec map[string]any, key string) (map[string]any, bool, error) {
	v, ok := sec[key]
	if !ok || v == nil {
		return nil, false, nil
	}

	if m, ok := v.(map[string]any); ok {
		return m, true, nil
	}

	// Some YAML decoders may produce map[any]any
	if m2, ok := v.(map[any]any); ok {
		out := make(map[string]any, len(m2))
		for k, vv := range m2 {
			ks, ok := k.(string)
			if !ok {
				return nil, false, fmt.Errorf("key %s must be map/object with string keys", key)
			}
			out[ks] = vv
		}
		return out, true, nil
	}

	return nil, false, fmt.Errorf("key %s must be map/object", key)
}

// AsDuration parses duration from YAML values.
// Accepts string (e.g. "10s") or time.Duration.
func AsDuration(v any) (time.Duration, error) {
	if v == nil {
		return 0, nil
	}

	switch x := v.(type) {
	case time.Duration:
		return x, nil
	case string:
		d, err := time.ParseDuration(x)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", x, err)
		}
		return d, nil
	default:
		return 0, fmt.Errorf("must be duration string")
	}
}
