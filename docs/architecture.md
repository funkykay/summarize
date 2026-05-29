# Architecture

## Directory structure

```text
cmd/summarize/
  main.go                 Program entry point.

internal/buildinfo/
  buildinfo.go            Build-time/runtime version value.

internal/cli/
  cli.go                  CLI parsing, command routing, command-specific option handling.

internal/config/
  config.go               Layered JSON configuration, dotted-path access, merge logic.

internal/ignore/
  ignore.go               Gitignore-like pattern matching and prune matching.

internal/initconfig/
  initconfig.go           Initial `summarize.json` generation rules.

internal/profile/
  profile.go              Profile normalization and profile layer application.

internal/selection/
  selection.go            File-level include/exclude selection.

internal/summary/
  summary.go              Recursive traversal and output generation.

internal/update/
  update.go               GitHub Release lookup, version comparison, binary download.

test/
  *_test.go               Integration-oriented tests and test helpers.
```

The project uses Go `internal` packages to keep implementation details private to the module.

## Entry point

`cmd/summarize/main.go` contains only process wiring:

```go
os.Exit(cli.New(os.Stdout, os.Stderr).Run(os.Args[1:]))
```

This keeps command execution testable because `internal/cli.App` accepts stdout and stderr writers.

## `internal/cli`

Responsibilities:

- Parse global options before a command name.
- Resolve `--base-dir` to an existing absolute directory.
- Track dry-run mode for the default summary command.
- Route commands.
- Convert returned errors into stderr output and exit code `1`.
- Call the summary, init, update, and version packages.

Supported commands:

- No command: create a summary.
- `init`: create `summarize.json`.
- `update`: check/download a newer standalone binary.
- `version`: print the version string.

The parser is hand-written instead of using a CLI framework. Given the small command surface, this avoids external dependencies and makes option handling explicit.

## `internal/config`

Core configuration abstraction.

Important types:

- `Dict`: `map[string]any` alias used for JSON-like data.
- `Config`: base config plus a stack of layers.
- `Layer`: one overlay with ID, name, path, and data.
- `LayerInfo`: public layer metadata.

Responsibilities:

- Load JSON files.
- Treat missing config files as empty objects.
- Reject JSON roots that are not objects.
- Push and pop layers.
- Merge layers deterministically.
- Support dotted-path access through `Get`, `Require`, and `Has`.

Merge behavior:

- Objects are recursively merged.
- Arrays are concatenated.
- Scalars are replaced by the overlay.

Layer IDs are generated as `layer-000001`, `layer-000002`, and so on.

## `internal/ignore`

Implements the pattern matcher used by pruning and selection.

Responsibilities:

- Normalize paths relative to the traversal base directory.
- Match anchored patterns such as `/go.mod`.
- Match directory patterns such as `build/`.
- Match wildcard patterns such as `*.log`.
- Evaluate prune negations through `IsPruned`.

The matcher translates patterns to regular expressions internally. It supports `*`, `?`, and bracket expressions.

## `internal/selection`

Turns the effective config into a file-selection decision.

Responsibilities:

- Normalize `selection_mode`.
- Normalize include and exclude pattern arrays.
- Decide whether a file should be printed.

Modes:

- `include_all`: include everything unless excluded, with include overriding exclude.
- `exclude_all`: exclude everything unless included, with exclude overriding include.

Selection uses descendant-aware matching so a directory pattern such as `src/` can include or exclude files below that directory.

## `internal/profile`

Applies a named profile from the current effective configuration.

Responsibilities:

- Read `profiles.<name>` from the merged config.
- Accept only valid profile keys.
- Normalize `selection_mode`.
- Normalize pattern arrays.
- Push a synthetic config layer named `profile:<name>`.

Unknown profiles are ignored. Profiles with no valid data are ignored.

## `internal/initconfig`

Builds a starter configuration for `summarize init`.

The current rule type is `PruneFolderIfExistRule`:

- If a folder exists, add that folder to `prune`.
- If any trigger path exists, add that folder to `prune`.
- Optionally use a custom prune entry instead of the folder name.

Current rules:

1. Add `node_modules/` when `node_modules/` or `package.json` exists.
2. Add `.git/` when `.git/` or `.gitignore` exists.

## `internal/summary`

Performs the main traversal.

Responsibilities:

- Resolve the absolute base directory.
- Load root configuration.
- Apply the selected profile.
- Traverse directories recursively.
- Push nested `summarize.json` layers when entering subdirectories.
- Pop nested layers when leaving subdirectories.
- Sort directory entries by name.
- Apply prune rules before recursion.
- Apply selection rules before file output.
- Print file blocks.
- Print path-only dry-run output when requested.
- Detect non-UTF-8 content and print a binary placeholder.
- Convert permission-denied directory reads into an output marker instead of failing the whole run.

## `internal/update`

Handles standalone binary updates.

Responsibilities:

- Determine the repository slug.
- Fetch the latest GitHub Release.
- Normalize and compare versions.
- Select the release asset for the current OS and architecture.
- Download the asset to `summarize.new` next to the current binary.
- Make the downloaded file executable.
- Print the command needed for manual replacement.

The update package is independent of `summarize.json` and traversal.

## Test architecture

Tests live in package `summarize_test`, outside the internal packages they exercise through the public CLI path or through package APIs where needed.

`test/testutil_test.go` builds a test binary once in `TestMain` and then runs commands against temporary file trees. This makes tests close to real CLI behavior, including stdout, stderr, working directory, and exit code handling.


