package summarize_test

import "testing"

func TestExcludeAllRequiresExplicitInclude(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"src/"},
		}).
		File("src/app.py", "print('app')\n").
		File("src/nested/util.py", "print('util')\n").
		File("tests/test_app.py", "print('test')\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"src/app.py", "src/nested/util.py"})
	validator.AssertNotContainsFile(t, "tests/test_app.py")
	validator.AssertFileContent(t, "src/app.py", "print('app')\n")
	validator.AssertFileContent(t, "src/nested/util.py", "print('util')\n")
}

func TestIncludeAllExcludesMatchingPaths(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"exclude":        []string{"src/"},
		}).
		File("src/app.py", "print('app')\n").
		File("docs/readme.md", "hello\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"docs/readme.md", "summarize.json"})
	validator.AssertNotContainsFile(t, "src/app.py")
	validator.AssertFileContent(t, "docs/readme.md", "hello\n")
}

func TestIncludeReincludesMatchingPathInIncludeAllMode(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"exclude":        []string{"*.log"},
			"include":        []string{"important.log"},
		}).
		File("app.log", "skip\n").
		File("important.log", "show\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"important.log", "summarize.json"})
	validator.AssertNotContainsFile(t, "app.log")
	validator.AssertFileContent(t, "important.log", "show\n")
}

func TestExcludeFiltersIncludedSubtreeInExcludeAllMode(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"src/"},
			"exclude":        []string{"src/generated/"},
		}).
		File("src/app.py", "print('app')\n").
		File("src/generated/client.py", "print('generated')\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"src/app.py"})
	validator.AssertNotContainsFile(t, "src/generated/client.py")
	validator.AssertFileContent(t, "src/app.py", "print('app')\n")
}

func TestNestedSummarizeJSONIsLoadedEvenWhenParentDirectoryIsNotSelected(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"visible.txt"},
		}).
		File("visible.txt", "visible\n").
		JSONFile("secret/summarize.json", map[string]any{
			"selection_mode": "include_all",
			"exclude":        []string{"skip.txt"},
		}).
		File("secret/revealed.txt", "revealed\n").
		File("secret/skip.txt", "skip\n").
		Create(t, root)

	output := runSummarize(t, root)
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{
		"secret/revealed.txt",
		"secret/summarize.json",
		"visible.txt",
	})
	validator.AssertNotContainsFile(t, "secret/skip.txt")
	validator.AssertFileContent(t, "secret/revealed.txt", "revealed\n")
	validator.AssertFileContent(t, "visible.txt", "visible\n")
}
