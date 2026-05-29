package initconfig

import (
	"os"
	"path/filepath"
)

type Config struct {
	SelectionMode string   `json:"selection_mode"`
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	Prune         []string `json:"prune"`
}

type Rule interface {
	Apply(baseDir string) []string
}

type PruneFolderIfExistRule struct {
	FolderName   string
	PruneEntry   string
	TriggerPaths []string
}

func (r PruneFolderIfExistRule) Apply(baseDir string) []string {
	if !r.shouldAddPrune(baseDir) {
		return nil
	}

	if r.PruneEntry != "" {
		return []string{r.PruneEntry}
	}

	return []string{r.FolderName + "/"}
}

func (r PruneFolderIfExistRule) shouldAddPrune(baseDir string) bool {
	if info, err := os.Stat(join(baseDir, r.FolderName)); err == nil && info.IsDir() {
		return true
	}

	for _, triggerPath := range r.TriggerPaths {
		if _, err := os.Stat(join(baseDir, triggerPath)); err == nil {
			return true
		}
	}

	return false
}

func BuildInitialConfig(baseDir string) Config {
	rules := []Rule{
		PruneFolderIfExistRule{
			FolderName:   "node_modules",
			TriggerPaths: []string{"package.json"},
		},
		PruneFolderIfExistRule{
			FolderName:   ".git",
			TriggerPaths: []string{".gitignore"},
		},
	}

	prune := make([]string, 0, len(rules))
	seen := map[string]struct{}{}

	for _, rule := range rules {
		for _, pruneEntry := range rule.Apply(baseDir) {
			if _, exists := seen[pruneEntry]; exists {
				continue
			}
			seen[pruneEntry] = struct{}{}
			prune = append(prune, pruneEntry)
		}
	}

	return Config{
		SelectionMode: "include_all",
		Include:       []string{},
		Exclude:       []string{},
		Prune:         prune,
	}

}

func join(baseDir string, name string) string {
	return filepath.Join(baseDir, name)
}
