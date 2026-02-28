package cmd

import (
	"fmt"
	"strings"
	"time"
)

func parseTags(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	tags := make(map[string]string, len(values))
	for _, raw := range values {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag %q; expected key=value", raw)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("invalid tag %q; empty key", raw)
		}

		tags[key] = val
	}

	return tags, nil
}

func parseOptionalTime(raw string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("invalid time %q, expected RFC3339", raw)
	}

	return &t, nil
}
