# Overview

## What summarize does

`summarize` serializes the contents of a directory tree into a stable plain-text representation.

Each included file is emitted as a block:

```text
=== relative/path/to/file ===
<file content>
```

The traversal order is deterministic: entries inside a directory are sorted by name before they are processed. A dry run can print only the selected file paths without reading file contents into the final output.

## Why it exists

The output is intentionally simple so it can be consumed by humans, scripts, and AI tools without a custom binary format or complex parser.

Common use cases:

- Create a text snapshot of a codebase.
- Feed selected project files into an AI assistant.
- Compare generated project summaries across runs.
- Check which files would be exported before generating the full output.
- Produce reproducible plain-text input for automation.

## Scope

`summarize` handles:

- Recursive directory traversal.
- Deterministic file ordering.
- JSON-based configuration through `summarize.json`.
- Directory-level pruning.
- File-level include and exclude selection.
- Named profile overlays.
- Nested configuration files in subdirectories.
- Dry-run listing of included file paths without file contents.
- Standalone binary update checks against GitHub Releases.

`summarize` does not handle:

- Semantic analysis of file contents.
- File content transformation.
- Syntax highlighting or rich output.
- Incremental cache-based summaries.
- Interactive prompts.
- Writing the summary directly to an output file.
- Windows standalone binary updates.

## Technical baseline

| Area | Current state |
|---|---|
| Language | Go |
| Go version | 1.22 |
| Module path | `github.com/funkykay/summarize` |
| Runtime dependencies | Go standard library only |
| Entry point | `cmd/summarize/main.go` |
| Configuration file | `summarize.json` |
| Test command | `go test ./...` |
| Build command | `go build -o bin/summarize ./cmd/summarize` |


