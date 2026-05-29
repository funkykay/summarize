# Development

## Requirements

- Go 1.22
- A POSIX-like shell for the examples below

The current `go.mod` has no external module dependencies.

## Common commands

Run tests:

```bash
go test ./...
```

Build the CLI:

```bash
go build -o bin/summarize ./cmd/summarize
```

Run from source:

```bash
go run ./cmd/summarize --base-dir .
```

Preview selected files without printing contents:

```bash
go run ./cmd/summarize --base-dir . --dry-run
```

Print version from source:

```bash
go run ./cmd/summarize version
```

## Project layout

```text
cmd/summarize/      Main package for the CLI binary.
internal/           Implementation packages.
test/               Integration-oriented tests and helpers.
go.mod              Module definition.
summarize.json      Project-local summarizer config.
```

## Testing approach

Tests in `test/` build an actual `summarize` test binary once in `TestMain` and then execute it through `os/exec`.

This verifies behavior close to real usage:

- Process exit code.
- stdout and stderr separation.
- Current working directory behavior.
- File tree traversal through the real filesystem.
- CLI option parsing.

Temporary file trees are created with helper methods in `test/testutil_test.go`.

## Current test coverage

| File | Area |
|---|---|
| `test/cli_test.go` | Version command, base directory handling, profile application, invalid base directory errors. |
| `test/dry_run_test.go` | Dry-run path-only output, profile and nested config behavior, command argument rejection. |
| `test/excludes_test.go` | Layered prune rules, negation, anchored patterns, nested configs, directory pruning. |
| `test/initialize_test.go` | `init` command behavior and automatic prune detection. |
| `test/profiles_test.go` | Profile include, exclude, prune, and selection mode overlays. |
| `test/selection_test.go` | `include_all`, `exclude_all`, include/exclude precedence, nested config loading. |
| `test/testutil_test.go` | Test binary build, process execution, file tree helpers, output validation. |

The current export does not contain dedicated tests for `internal/update` or direct unit tests for every `internal/config` edge case.

## Code conventions

- Use tabs for Go indentation, matching `gofmt`.
- Keep packages small and responsibility-focused.
- Prefer explicit error returns over process exits inside internal packages.
- Keep stdout and stderr injectable for testability.
- Do not add runtime dependencies unless the dependency removes more complexity than it introduces.
- Keep output format stable unless a breaking change is intentional.

## Adding a command

A new command currently requires changes in `internal/cli/cli.go`:

1. Extend the `switch command` in `App.Run`.
2. Add a `run<Command>` method if command-specific parsing is needed.
3. Keep command output on `a.stdout` and errors as returned `error` values.
4. Decide whether existing global options such as `--base-dir`, `--profile`, and `--dry-run` should apply to the new command or be rejected after the command name.
5. Add CLI-level tests through the built test binary.

## Adding a config key

A new config key usually requires changes in more than one package:

1. Define how the key is represented in JSON.
2. Add normalization close to where the key is consumed.
3. Decide merge behavior. Existing merge behavior is type-based: objects merge, arrays concatenate, scalars replace.
4. Update profile handling if the key should be valid inside profiles.
5. Add tests for root config, profile config, and nested config if applicable.
6. Update documentation.

## Adding an init rule

`internal/initconfig` is rule-based. Add another `Rule` implementation or another `PruneFolderIfExistRule` entry in `BuildInitialConfig`.

Rules should remain deterministic. Preserve stable output order and deduplicate generated prune entries.

## Formatting

Use standard Go tooling:

```bash
gofmt -w <files>
go test ./...
```


