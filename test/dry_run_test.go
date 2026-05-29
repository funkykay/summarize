package summarize_test

import (
	"strings"
	"testing"
)

func TestDryRunOutputsOnlyExportedFilePaths(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"include":        []string{},
			"exclude":        []string{"summarize.json", "*.log"},
			"prune":          []string{"ignored/"},
		}).
		File("alpha.txt", "Alpha\n").
		File("beta.log", "Beta\n").
		File("ignored/secret.txt", "Secret\n").
		File("nested/gamma.txt", "Gamma\n").
		Create(t, root)

	result := runCLIProcess(t, root, "--dry-run")

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}

	expected := strings.Join([]string{
		"alpha.txt",
		"nested/gamma.txt",
		"",
	}, "\n")
	if result.Stdout != expected {
		t.Fatalf("unexpected stdout: expected %q, got %q", expected, result.Stdout)
	}
	if strings.Contains(result.Stdout, "===") {
		t.Fatalf("expected dry-run output without file headers, got %q", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Alpha") || strings.Contains(result.Stdout, "Gamma") {
		t.Fatalf("expected dry-run output without file contents, got %q", result.Stdout)
	}
}

func TestDryRunUsesProfileAndNestedConfigLayers(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"src/"},
			"profiles": map[string]any{
				"ci": map[string]any{
					"include": []string{"docs/"},
					"exclude": []string{"src/generated/"},
					"prune":   []string{"reports/"},
				},
			},
		}).
		File("src/app.go", "package main\n").
		File("src/generated/client.go", "package generated\n").
		File("docs/readme.md", "# Docs\n").
		File("reports/junit.xml", "<testsuite />\n").
		JSONFile("tools/summarize.json", map[string]any{
			"selection_mode": "include_all",
			"exclude":        []string{"skip.txt"},
		}).
		File("tools/run.sh", "#!/bin/sh\n").
		File("tools/skip.txt", "skip\n").
		Create(t, root)

	result := runCLIProcess(t, root, "--dry-run", "--profile", "ci")

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", result.ExitCode, result.Stderr)
	}
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}

	expected := strings.Join([]string{
		"docs/readme.md",
		"src/app.go",
		"tools/run.sh",
		"tools/summarize.json",
		"",
	}, "\n")
	if result.Stdout != expected {
		t.Fatalf("unexpected stdout: expected %q, got %q", expected, result.Stdout)
	}
	if strings.Contains(result.Stdout, "package main") || strings.Contains(result.Stdout, "# Docs") {
		t.Fatalf("expected dry-run output without file contents, got %q", result.Stdout)
	}
}

func TestDryRunIsRejectedForCommands(t *testing.T) {
	root := t.TempDir()

	result := runCLIProcess(t, root, "version", "--dry-run")

	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Stdout != "" {
		t.Fatalf("expected empty stdout, got %q", result.Stdout)
	}
	if !strings.Contains(result.Stderr, "unexpected argument for version: --dry-run") {
		t.Fatalf("expected unexpected argument error, got %q", result.Stderr)
	}
}
