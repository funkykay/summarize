package summary

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"

	"github.com/funkykay/summarize/internal/config"
	"github.com/funkykay/summarize/internal/ignore"
	"github.com/funkykay/summarize/internal/profile"
	"github.com/funkykay/summarize/internal/selection"
)

func Create(profileName string, baseDir string, writer io.Writer) error {
	resolvedBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}

	cfg, err := config.FromFile(filepath.Join(resolvedBaseDir, "summarize.json"))
	if err != nil {
		return err
	}
	profile.ApplyLayer(cfg, profileName)

	return Directory(resolvedBaseDir, cfg, resolvedBaseDir, writer, false)
}

func Directory(directory string, cfg *config.Config, basePath string, writer io.Writer, loadConfigLayer bool) error {
	layerPushed := false

	if loadConfigLayer {
		_, pushed, err := cfg.PushLayer(filepath.Join(directory, "summarize.json"))
		if err != nil {
			return err
		}
		layerPushed = pushed
	}

	defer func() {
		if layerPushed {
			_, _ = cfg.PopLayer()
		}
	}()

	entries, err := os.ReadDir(directory)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			relativePath, relErr := filepath.Rel(basePath, directory)
			if relErr != nil {
				relativePath = directory
			}
			fmt.Fprintf(writer, "[Access denied: %s]\n", filepath.ToSlash(relativePath))
			return nil
		}
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	prunePatterns := normalizePatterns(cfg.Get("prune", []any{}))
	traversalSelection := selection.FromConfig(cfg)

	for _, entry := range entries {
		itemPath := filepath.Join(directory, entry.Name())
		if ignore.IsPruned(itemPath, basePath, prunePatterns) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := Directory(itemPath, cfg, basePath, writer, true); err != nil {
				return err
			}
			continue
		}

		if !traversalSelection.AllowsPath(itemPath, basePath) {
			continue
		}

		relativePath, err := filepath.Rel(basePath, itemPath)
		if err != nil {
			relativePath = itemPath
		}

		fmt.Fprintf(writer, "=== %s ===\n", filepath.ToSlash(relativePath))
		printFileContent(writer, itemPath)
		fmt.Fprintln(writer)
	}

	return nil
}

func printFileContent(writer io.Writer, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(writer, "[Error reading file: %s]\n", err)
		return
	}

	if !utf8.Valid(content) {
		fmt.Fprintln(writer, "[Binary file - content not displayable]")
		return
	}

	fmt.Fprintln(writer, string(content))
}

func normalizePatterns(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		patterns := make([]string, 0, len(typed))
		for _, entry := range typed {
			text, ok := entry.(string)
			if ok {
				patterns = append(patterns, text)
			}
		}
		return patterns
	default:
		return nil
	}
}
