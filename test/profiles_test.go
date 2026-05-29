package summarize_test

import "testing"

func TestProfilePruneIsAppliedWhenProfileIsSelected(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"profiles": map[string]any{
				"test": map[string]any{"prune": []string{"tests/"}},
			},
		}).
		File("main.py", "print('main')\n").
		File("tests/test_main.py", "print('test')\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "test")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"main.py", "summarize.json"})
	validator.AssertNotContainsFile(t, "tests/test_main.py")
	validator.AssertFileContent(t, "main.py", "print('main')\n")
}

func TestProfilePruneIsMergedWithDirectoryPrune(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"prune": []string{"*.tmp"},
			"profiles": map[string]any{
				"ci": map[string]any{"prune": []string{"reports/"}},
			},
		}).
		File("keep.txt", "keep\n").
		File("cache.tmp", "tmp\n").
		File("reports/junit.xml", "xml\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "ci")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"keep.txt", "summarize.json"})
	validator.AssertNotContainsFile(t, "cache.tmp")
	validator.AssertNotContainsFile(t, "reports/junit.xml")
	validator.AssertFileContent(t, "keep.txt", "keep\n")
}

func TestProfileIncludeIsAppliedWhenProfileIsSelected(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"main.py"},
			"profiles": map[string]any{
				"test": map[string]any{"include": []string{"tests/"}},
			},
		}).
		File("main.py", "print('main')\n").
		File("tests/test_main.py", "print('test')\n").
		File("docs/readme.md", "hidden\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "test")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"main.py", "tests/test_main.py"})
	validator.AssertNotContainsFile(t, "docs/readme.md")
	validator.AssertFileContent(t, "main.py", "print('main')\n")
	validator.AssertFileContent(t, "tests/test_main.py", "print('test')\n")
}

func TestProfileExcludeIsMergedWithDirectoryInclude(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"src/"},
			"profiles": map[string]any{
				"ci": map[string]any{"exclude": []string{"src/generated/"}},
			},
		}).
		File("src/app.py", "print('app')\n").
		File("src/generated/client.py", "print('generated')\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "ci")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"src/app.py"})
	validator.AssertNotContainsFile(t, "src/generated/client.py")
	validator.AssertFileContent(t, "src/app.py", "print('app')\n")
}

func TestProfileSelectionModeOverridesDirectorySelectionModeToIncludeAll(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"main.py"},
			"profiles": map[string]any{
				"test": map[string]any{"selection_mode": "include_all"},
			},
		}).
		File("main.py", "print('main')\n").
		File("hidden.txt", "now visible\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "test")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"hidden.txt", "main.py", "summarize.json"})
	validator.AssertFileContent(t, "hidden.txt", "now visible\n")
	validator.AssertFileContent(t, "main.py", "print('main')\n")
}

func TestProfileSelectionModeOverridesDirectorySelectionModeToExcludeAll(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"profiles": map[string]any{
				"minimal": map[string]any{
					"selection_mode": "exclude_all",
					"include":        []string{"visible.txt"},
				},
			},
		}).
		File("visible.txt", "visible\n").
		File("hidden.txt", "hidden\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "minimal")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"visible.txt"})
	validator.AssertNotContainsFile(t, "hidden.txt")
	validator.AssertNotContainsFile(t, "summarize.json")
	validator.AssertFileContent(t, "visible.txt", "visible\n")
}

func TestInvalidProfileSelectionModeIsIgnored(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "exclude_all",
			"include":        []string{"main.py"},
			"profiles": map[string]any{
				"test": map[string]any{
					"selection_mode": "invalid",
					"include":        []string{"profile.txt"},
				},
			},
		}).
		File("main.py", "print('main')\n").
		File("profile.txt", "profile\n").
		File("hidden.txt", "hidden\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "test")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"main.py", "profile.txt"})
	validator.AssertNotContainsFile(t, "hidden.txt")
	validator.AssertNotContainsFile(t, "summarize.json")
	validator.AssertFileContent(t, "main.py", "print('main')\n")
	validator.AssertFileContent(t, "profile.txt", "profile\n")
}

func TestProfileSelectionModeKeepsPatternLayersAdditive(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"selection_mode": "include_all",
			"include":        []string{"docs/"},
			"exclude":        []string{"src/generated/"},
			"prune":          []string{"cache/"},
			"profiles": map[string]any{
				"ci": map[string]any{
					"selection_mode": "exclude_all",
					"include":        []string{"src/"},
					"exclude":        []string{"src/experimental/"},
					"prune":          []string{"reports/"},
				},
			},
		}).
		File("docs/readme.md", "docs\n").
		File("src/app.py", "print('app')\n").
		File("src/generated/client.py", "print('generated')\n").
		File("src/experimental/demo.py", "print('experimental')\n").
		File("cache/data.txt", "cache\n").
		File("reports/junit.xml", "xml\n").
		File("hidden.txt", "hidden\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "ci")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"docs/readme.md", "src/app.py"})
	validator.AssertNotContainsFile(t, "src/generated/client.py")
	validator.AssertNotContainsFile(t, "src/experimental/demo.py")
	validator.AssertNotContainsFile(t, "cache/data.txt")
	validator.AssertNotContainsFile(t, "reports/junit.xml")
	validator.AssertNotContainsFile(t, "hidden.txt")
	validator.AssertFileContent(t, "docs/readme.md", "docs\n")
	validator.AssertFileContent(t, "src/app.py", "print('app')\n")
}

func TestUnknownProfileDoesNotChangeOutput(t *testing.T) {
	root := t.TempDir()
	newFileTree().
		JSONFile("summarize.json", map[string]any{
			"profiles": map[string]any{
				"prod": map[string]any{"prune": []string{"prod-only.txt"}},
			},
		}).
		File("visible.txt", "visible\n").
		File("prod-only.txt", "still visible\n").
		Create(t, root)

	output := runSummarize(t, root, "--profile", "missing")
	validator := newSummarizeOutputValidator(t, output)

	validator.AssertPaths(t, []string{"prod-only.txt", "summarize.json", "visible.txt"})
	validator.AssertFileContent(t, "prod-only.txt", "still visible\n")
	validator.AssertFileContent(t, "visible.txt", "visible\n")
}
