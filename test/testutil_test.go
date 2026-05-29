package summarize_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
)

var summarizeBinary string

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "summarize-tests-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	binaryPath := filepath.Join(tempDir, "summarize-test")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "failed to determine test file path")
		_ = os.RemoveAll(tempDir)
		os.Exit(1)
	}
	projectRoot := filepath.Dir(filepath.Dir(file))

	build := exec.Command("go", "build", "-o", binaryPath, "./cmd/summarize")
	build.Dir = projectRoot
	build.Env = os.Environ()
	output, err := build.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build summarize test binary: %v\n%s", err, output)
		_ = os.RemoveAll(tempDir)
		os.Exit(1)
	}

	summarizeBinary = binaryPath
	code := m.Run()
	_ = os.RemoveAll(tempDir)
	os.Exit(code)
}

type processResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runCLIProcess(t *testing.T, cwd string, args ...string) processResult {
	t.Helper()

	cmd := exec.Command(summarizeBinary, args...)
	cmd.Dir = cwd

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatalf("run summarize: %v", err)
		}
		exitCode = exitError.ExitCode()
	}

	return processResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

func runSummarize(t *testing.T, cwd string, args ...string) string {
	t.Helper()

	result := runCLIProcess(t, cwd, args...)
	if result.ExitCode != 0 {
		t.Fatalf("summarize failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	return result.Stdout
}

func runInitProcess(t *testing.T, cwd string, args ...string) processResult {
	t.Helper()

	allArgs := append([]string{"init"}, args...)
	return runCLIProcess(t, cwd, allArgs...)
}

type fileTree struct {
	entries []treeEntry
}

type treeEntry interface {
	create(root string) error
}

type fileEntry struct {
	path    string
	content string
}

type jsonFileEntry struct {
	path string
	data map[string]any
}

type binaryFileEntry struct {
	path string
	size int
}

type directoryEntry struct {
	path string
}

func newFileTree() *fileTree {
	return &fileTree{}
}

func (tree *fileTree) File(path string, content string) *fileTree {
	tree.entries = append(tree.entries, fileEntry{path: path, content: content})
	return tree
}

func (tree *fileTree) JSONFile(path string, data map[string]any) *fileTree {
	tree.entries = append(tree.entries, jsonFileEntry{path: path, data: data})
	return tree
}

func (tree *fileTree) BinaryFile(path string, size int) *fileTree {
	if size < 0 {
		panic("binary file size must be >= 0")
	}

	tree.entries = append(tree.entries, binaryFileEntry{path: path, size: size})
	return tree
}

func (tree *fileTree) Directory(path string) *fileTree {
	tree.entries = append(tree.entries, directoryEntry{path: path})
	return tree
}

func (tree *fileTree) Create(t *testing.T, root string) string {
	t.Helper()

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("create root: %v", err)
	}

	for _, entry := range tree.entries {
		if err := entry.create(root); err != nil {
			t.Fatalf("create file tree: %v", err)
		}
	}

	return root
}

func (entry fileEntry) create(root string) error {
	path := filepath.Join(root, filepath.FromSlash(entry.path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(entry.content), 0o644)
}

func (entry jsonFileEntry) create(root string) error {
	content, err := json.MarshalIndent(entry.data, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')

	path := filepath.Join(root, filepath.FromSlash(entry.path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0o644)
}

func (entry binaryFileEntry) create(root string) error {
	path := filepath.Join(root, filepath.FromSlash(entry.path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, bytes.Repeat([]byte{0xff}, entry.size), 0o644)
}

func (entry directoryEntry) create(root string) error {
	return os.MkdirAll(filepath.Join(root, filepath.FromSlash(entry.path)), 0o755)
}

type summarizeOutputValidator struct {
	paths []string
	files map[string]string
}

var headerPattern = regexp.MustCompile(`(?m)^=== (.+) ===\n`)

func newSummarizeOutputValidator(t *testing.T, output string) summarizeOutputValidator {
	t.Helper()

	matches := headerPattern.FindAllStringSubmatchIndex(output, -1)
	files := make(map[string]string, len(matches))
	paths := make([]string, 0, len(matches))

	for index, match := range matches {
		path := output[match[2]:match[3]]
		bodyStart := match[1]
		bodyEnd := len(output)
		if index+1 < len(matches) {
			bodyEnd = matches[index+1][0]
		}

		body := output[bodyStart:bodyEnd]
		if !strings.HasSuffix(body, "\n\n") {
			t.Fatalf("invalid summarize output block for %q", path)
		}
		if _, exists := files[path]; exists {
			t.Fatalf("duplicate file entry %q in summarize output", path)
		}

		paths = append(paths, path)
		files[path] = strings.TrimSuffix(body, "\n\n")
	}

	return summarizeOutputValidator{paths: paths, files: files}
}

func (validator summarizeOutputValidator) AssertPaths(t *testing.T, expected []string) {
	t.Helper()

	if !slices.Equal(validator.paths, expected) {
		t.Fatalf("unexpected summarized file list: expected %v, got %v", expected, validator.paths)
	}
}

func (validator summarizeOutputValidator) AssertContainsFile(t *testing.T, path string) {
	t.Helper()

	if _, exists := validator.files[path]; !exists {
		t.Fatalf("expected file %q to be present in summarize output", path)
	}
}

func (validator summarizeOutputValidator) AssertNotContainsFile(t *testing.T, path string) {
	t.Helper()

	if _, exists := validator.files[path]; exists {
		t.Fatalf("expected file %q to be absent from summarize output", path)
	}
}

func (validator summarizeOutputValidator) AssertFileContent(t *testing.T, path string, expected string) {
	t.Helper()

	actual, exists := validator.files[path]
	if !exists {
		t.Fatalf("expected file %q to be present in summarize output", path)
	}
	if actual != expected {
		t.Fatalf("unexpected content for %q: expected %q, got %q", path, expected, actual)
	}
}

type summarizeConfig struct {
	SelectionMode string   `json:"selection_mode"`
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	Prune         []string `json:"prune"`
}

func readSummarizeConfig(t *testing.T, path string) summarizeConfig {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read summarize config: %v", err)
	}

	var cfg summarizeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("decode summarize config: %v", err)
	}

	return cfg
}

func assertEqual[T any](t *testing.T, expected T, actual T) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
