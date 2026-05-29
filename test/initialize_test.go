package summarize_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesEmptyPruneWhenNoRuleMatches(t *testing.T) {
	root := t.TempDir()

	result := runInitProcess(t, root)

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
	configPath := filepath.Join(root, "summarize.json")
	if result.Stdout != "Created "+configPath+"\n" {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	assertEqual(t, expectedConfig(nil), readSummarizeConfig(t, configPath))
}

func TestInitAddsGitPruneWhenGitDirectoryExists(t *testing.T) {
	root := t.TempDir()
	newFileTree().Directory(".git").Create(t, root)

	result := runInitProcess(t, root)

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	assertEqual(t, expectedConfig([]string{".git/"}), readSummarizeConfig(t, filepath.Join(root, "summarize.json")))
}

func TestInitAddsNodeModulesPruneWhenNodeModulesDirectoryExists(t *testing.T) {
	root := t.TempDir()
	newFileTree().Directory("node_modules").Create(t, root)

	result := runInitProcess(t, root)

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	assertEqual(t, expectedConfig([]string{"node_modules/"}), readSummarizeConfig(t, filepath.Join(root, "summarize.json")))
}

func TestInitAddsNodeModulesPruneWhenPackageJSONExists(t *testing.T) {
	root := t.TempDir()
	newFileTree().File("package.json", `{ "name": "demo" }`+"\n").Create(t, root)

	result := runInitProcess(t, root)

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	assertEqual(t, expectedConfig([]string{"node_modules/"}), readSummarizeConfig(t, filepath.Join(root, "summarize.json")))
}

func TestInitDeduplicatesNodeModulesPruneWhenFolderAndPackageJSONExist(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		Directory("node_modules").
		File("package.json", `{ "name": "demo" }`+"\n").
		Create(t, root)

	result := runInitProcess(t, root)

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	assertEqual(t, expectedConfig([]string{"node_modules/"}), readSummarizeConfig(t, filepath.Join(root, "summarize.json")))
}

func TestInitAddsAllMatchingPruneInRuleOrder(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		Directory(".git").
		File("package.json", `{ "name": "demo" }`+"\n").
		Create(t, root)

	result := runInitProcess(t, root)

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	assertEqual(t, expectedConfig([]string{"node_modules/", ".git/"}), readSummarizeConfig(t, filepath.Join(root, "summarize.json")))
}

func TestInitFailsWhenSummarizeJSONAlreadyExists(t *testing.T) {
	root := t.TempDir()
	newFileTree().JSONFile("summarize.json", map[string]any{"prune": []string{".git/"}}).Create(t, root)

	result := runInitProcess(t, root)

	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Stdout != "" {
		t.Fatalf("expected empty stdout, got %q", result.Stdout)
	}
	if !strings.Contains(result.Stderr, "Configuration file already exists:") {
		t.Fatalf("expected existing config error, got %q", result.Stderr)
	}
	if _, err := os.Stat(filepath.Join(root, "summarize.json")); err != nil {
		t.Fatalf("expected existing config to remain: %v", err)
	}
}

func expectedConfig(prune []string) summarizeConfig {
	if prune == nil {
		prune = []string{}
	}

	return summarizeConfig{
		SelectionMode: "include_all",
		Include:       []string{},
		Exclude:       []string{},
		Prune:         prune,
	}
}
