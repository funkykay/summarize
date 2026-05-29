package ignore

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GitignoreMatch(pattern string, pathString string, isDir bool) bool {
	normalizedPattern := strings.TrimSpace(pattern)
	if normalizedPattern == "" || strings.HasPrefix(normalizedPattern, "#") {
		return false
	}

	if strings.HasSuffix(normalizedPattern, "/") {
		if !isDir {
			return false
		}
		normalizedPattern = strings.TrimSuffix(normalizedPattern, "/")
	}

	if strings.HasPrefix(normalizedPattern, "/") {
		anchored := strings.TrimPrefix(normalizedPattern, "/")
		return pathString == anchored || strings.HasPrefix(pathString, anchored+"/")
	}

	parts := strings.Split(pathString, "/")

	if fnmatch(pathString, normalizedPattern) {
		return true
	}

	for i := range parts {
		subpath := strings.Join(parts[i:], "/")
		if fnmatch(subpath, normalizedPattern) {
			return true
		}
		if fnmatch(parts[i], normalizedPattern) {
			return true
		}
	}

	return false
}

func GitignoreMatchWithDescendants(pattern string, pathString string, isDir bool) bool {
	if GitignoreMatch(pattern, pathString, isDir) {
		return true
	}

	rawParts := strings.Split(pathString, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if part != "" {
			parts = append(parts, part)
		}
	}

	upperBound := len(parts)
	if !isDir {
		upperBound = len(parts) - 1
	}

	for i := 1; i <= upperBound; i++ {
		ancestor := strings.Join(parts[:i], "/")
		if ancestor == pathString {
			continue
		}

		if GitignoreMatch(pattern, ancestor, true) {
			return true
		}
	}

	return false
}

func MatchesAnyPattern(patterns []string, path string, basePath string, matchDescendants bool) bool {
	pathString := relativeSlashPath(path, basePath)
	isDir := isDirectory(path)

	for _, pattern := range patterns {
		if matchDescendants {
			if GitignoreMatchWithDescendants(pattern, pathString, isDir) {
				return true
			}
			continue
		}

		if GitignoreMatch(pattern, pathString, isDir) {
			return true
		}
	}

	return false
}

func IsPruned(path string, basePath string, prunePatterns []string) bool {
	if len(prunePatterns) == 0 {
		return false
	}

	pathString := relativeSlashPath(path, basePath)
	isDir := isDirectory(path)
	pruned := false

	for _, rawPattern := range prunePatterns {
		pattern := strings.TrimSpace(rawPattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		if strings.HasPrefix(pattern, "!") {
			if GitignoreMatch(strings.TrimPrefix(pattern, "!"), pathString, isDir) {
				pruned = false
			}
			continue
		}

		if GitignoreMatch(pattern, pathString, isDir) {
			pruned = true
		}
	}

	return pruned
}

func relativeSlashPath(path string, basePath string) string {
	relativePath, err := filepath.Rel(basePath, path)
	if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		relativePath = path
	}

	return filepath.ToSlash(relativePath)
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fnmatch(name string, pattern string) bool {
	expression, err := regexp.Compile("^" + translatePattern(pattern) + "$")
	if err != nil {
		return false
	}

	return expression.MatchString(name)
}

func translatePattern(pattern string) string {
	var builder strings.Builder

	for i := 0; i < len(pattern); i++ {
		char := pattern[i]

		switch char {
		case '*':
			builder.WriteString(".*")
		case '?':
			builder.WriteByte('.')
		case '[':
			end := i + 1
			if end < len(pattern) && pattern[end] == '!' {
				end++
			}
			if end < len(pattern) && pattern[end] == ']' {
				end++
			}
			for end < len(pattern) && pattern[end] != ']' {
				end++
			}
			if end >= len(pattern) {
				builder.WriteString(regexp.QuoteMeta("["))
				continue
			}

			content := pattern[i+1 : end]
			if strings.HasPrefix(content, "!") {
				content = "^" + regexp.QuoteMeta(content[1:])
			} else {
				content = regexp.QuoteMeta(content)
			}
			content = strings.ReplaceAll(content, `\-`, `-`)
			builder.WriteString("[" + content + "]")
			i = end
		default:
			builder.WriteString(regexp.QuoteMeta(string(char)))
		}
	}

	return builder.String()
}
