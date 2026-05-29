package selection

import (
	"strings"

	"github.com/funkykay/summarize/internal/config"
	"github.com/funkykay/summarize/internal/ignore"
)

type Mode string

const (
	ModeIncludeAll Mode = "include_all"
	ModeExcludeAll Mode = "exclude_all"
)

type TraversalSelection struct {
	Mode    Mode
	Include []string
	Exclude []string
}

func ModeFromValue(value any) Mode {
	text, ok := value.(string)
	if ok {
		normalized := strings.ToLower(strings.TrimSpace(text))
		if normalized == string(ModeIncludeAll) {
			return ModeIncludeAll
		}
		if normalized == string(ModeExcludeAll) {
			return ModeExcludeAll
		}
	}

	return ModeIncludeAll
}

func FromConfig(cfg *config.Config) TraversalSelection {
	return TraversalSelection{
		Mode:    ModeFromValue(cfg.Get("selection_mode", nil)),
		Include: normalizePatterns(cfg.Get("include", []any{})),
		Exclude: normalizePatterns(cfg.Get("exclude", []any{})),
	}
}

func (s TraversalSelection) AllowsPath(path string, basePath string) bool {
	explicitlyIncluded := ignore.MatchesAnyPattern(s.Include, path, basePath, true)
	explicitlyExcluded := ignore.MatchesAnyPattern(s.Exclude, path, basePath, true)

	if s.Mode == ModeIncludeAll {
		return explicitlyIncluded || !explicitlyExcluded
	}

	return explicitlyIncluded && !explicitlyExcluded
}

func normalizePatterns(value any) []string {
	switch typed := value.(type) {
	case []string:
		patterns := make([]string, 0, len(typed))
		for _, entry := range typed {
			pattern := strings.TrimSpace(entry)
			if pattern != "" {
				patterns = append(patterns, pattern)
			}
		}
		return patterns
	case []any:
		patterns := make([]string, 0, len(typed))
		for _, entry := range typed {
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
	default:
		return nil
	}
}
