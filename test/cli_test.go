package summarize_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/funkykay/summarize/internal/buildinfo"
	"github.com/funkykay/summarize/internal/cli"
)

func TestVersionOutputsPackageVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.New(&stdout, &stderr).Run([]string{"version"})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if stdout.String() != buildinfo.Version {
		t.Fatalf("expected version %q, got %q", buildinfo.Version, stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestSummarizeUsesBaseDir(t *testing.T) {
	baseDir := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"include":        []string{},
			"exclude":        []string{"summarize.json"},
			"prune":          []string{},
		}).
		File("example.txt", "Hello\n").
		Create(t, baseDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.New(&stdout, &stderr).Run([]string{"--base-dir", baseDir})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if stdout.String() != "=== example.txt ===\nHello\n\n\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestSummarizeAppliesProfile(t *testing.T) {
	baseDir := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"include":        []string{},
			"exclude":        []string{"summarize.json"},
			"prune":          []string{},
			"profiles": map[string]any{
				"minimal": map[string]any{
					"exclude": []string{"ignored.txt"},
				},
			},
		}).
		File("included.txt", "Included\n").
		File("ignored.txt", "Ignored\n").
		Create(t, baseDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.New(&stdout, &stderr).Run([]string{"--base-dir", baseDir, "--profile", "minimal"})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if stdout.String() != "=== included.txt ===\nIncluded\n\n\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestInvalidBaseDirReturnsError(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.New(&stdout, &stderr).Run([]string{"--base-dir", missingDir})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.String() != "" {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(strings.ToLower(stderr.String()), "does not exist") {
		t.Fatalf("expected missing directory error, got %q", stderr.String())
	}
}
