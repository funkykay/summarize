package profile

import (
	"strings"

	"github.com/funkykay/summarize/internal/config"
)

func ApplyLayer(cfg *config.Config, profileName string) {
	if profileName == "" {
		return
	}

	merged := cfg.ToMap()
	profiles, ok := asMap(merged["profiles"])
	if !ok {
		return
	}

	profileData, ok := asMap(profiles[profileName])
	if !ok {
		return
	}

	layerData := config.Dict{}

	if selectionMode := normalizeSelectionMode(profileData["selection_mode"]); selectionMode != "" {
		layerData["selection_mode"] = selectionMode
	}

	for _, key := range []string{"prune", "include", "exclude"} {
		patterns := normalizePatterns(profileData[key])
		if len(patterns) > 0 {
			layerData[key] = patterns
		}
	}

	if len(layerData) == 0 {
		return
	}

	cfg.PushDataLayer(layerData, "profile:"+profileName, "<profile>")
}

func normalizeSelectionMode(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}

	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "include_all" || normalized == "exclude_all" {
		return normalized
	}

	return ""
}

func normalizePatterns(value any) []string {
	values, ok := value.([]any)
	if !ok {
		if typed, typedOK := value.([]string); typedOK {
			patterns := make([]string, 0, len(typed))
			for _, entry := range typed {
				pattern := strings.TrimSpace(entry)
				if pattern != "" {
					patterns = append(patterns, pattern)
				}
			}
			return patterns
		}
		return nil
	}

	patterns := make([]string, 0, len(values))
	for _, entry := range values {
		text, ok := entry.(string)
		if !ok {
			continue
		}

		pattern := strings.TrimSpace(text)
		if pattern != "" {
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

func asMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case config.Dict:
		return map[string]any(typed), true
	default:
		return nil, false
	}
}
