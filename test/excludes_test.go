package summarize_test

import "testing"

func TestAppliesLayeredSummarizeJSONPrune(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{"prune": []string{"root-secret.txt", "ignored/"}}).
		File("visible.txt", "visible\n").
		File("root-secret.txt", "hidden\n").
		File("ignored/never.txt", "nope\n").
		File("project/keep.txt", "keep\n").
		JSONFile("project/summarize.json", map[string]any{"prune": []string{"deep-secret.txt"}}).
		File("project/deep-secret.txt", "hidden too\n").
		File("project/nested/seen.txt", "seen\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{
		"project/keep.txt",
		"project/nested/seen.txt",
		"project/summarize.json",
		"summarize.json",
		"visible.txt",
	})
	validator.AssertNotContainsFile(t, "root-secret.txt")
	validator.AssertNotContainsFile(t, "ignored/never.txt")
	validator.AssertNotContainsFile(t, "project/deep-secret.txt")
	validator.AssertFileContent(t, "project/keep.txt", "keep\n")
	validator.AssertFileContent(t, "project/nested/seen.txt", "seen\n")
	validator.AssertFileContent(t, "visible.txt", "visible\n")
}

func TestNegationPatternReincludesFile(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{"prune": []string{"*.log", "!important.log"}}).
		File("app.log", "skip\n").
		File("important.log", "show\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"important.log", "summarize.json"})
	validator.AssertNotContainsFile(t, "app.log")
	validator.AssertFileContent(t, "important.log", "show\n")
}

func TestNestedSummarizeJSONOnlyAppliesBelowItsDirectory(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		File("hidden.txt", "root-visible").
		File("parent.txt", "parent").
		JSONFile("scope/summarize.json", map[string]any{"prune": []string{"hidden.txt"}}).
		File("scope/hidden.txt", "hidden").
		File("scope/nested/visible.txt", "visible").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{
		"hidden.txt",
		"parent.txt",
		"scope/nested/visible.txt",
		"scope/summarize.json",
	})
	validator.AssertNotContainsFile(t, "scope/hidden.txt")
	validator.AssertFileContent(t, "hidden.txt", "root-visible")
	validator.AssertFileContent(t, "parent.txt", "parent")
	validator.AssertFileContent(t, "scope/nested/visible.txt", "visible")
}

func TestAnchoredPatternOnlyMatchesFromRoot(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{"prune": []string{"/root-only.txt"}}).
		File("root-only.txt", "hidden\n").
		File("nested/root-only.txt", "visible\n").
		File("nested/keep.txt", "keep\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"nested/keep.txt", "nested/root-only.txt", "summarize.json"})
	validator.AssertNotContainsFile(t, "root-only.txt")
	validator.AssertFileContent(t, "nested/root-only.txt", "visible\n")
	validator.AssertFileContent(t, "nested/keep.txt", "keep\n")
}

func TestDirectoryPatternWithoutAnchorMatchesNestedDirectories(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{"prune": []string{"build/"}}).
		File("build/root.txt", "skip root\n").
		File("src/build/generated.txt", "skip nested\n").
		File("src/app.py", "print('ok')\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"src/app.py", "summarize.json"})
	validator.AssertNotContainsFile(t, "build/root.txt")
	validator.AssertNotContainsFile(t, "src/build/generated.txt")
	validator.AssertFileContent(t, "src/app.py", "print('ok')\n")
}
