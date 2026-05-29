# summarize

`summarize` is a Go command-line tool that recursively serializes a directory tree into a stable, plain-text format.

It is designed for workflows where a complete, deterministic snapshot of relevant project files should be passed to another tool, for example an AI assistant, a code review process, or a text-processing pipeline.

## Features

- Recursively traverses a directory tree.
- Outputs files in deterministic lexicographic order.
- Emits every included file as a parseable block using `=== path ===` headers.
- Supports dry-run output for inspecting selected files without printing file contents.
- Supports layered JSON configuration through `summarize.json` files.
- Supports include, exclude, and prune patterns with gitignore-like matching.
- Supports named profiles as configuration overlays.
- Provides `init`, `version`, and standalone-binary `update` commands.
- Uses only the Go standard library at runtime.

## Installation

### Build from source

```bash
go build -o bin/summarize ./cmd/summarize
```

Then run the binary directly:

```bash
./bin/summarize --base-dir path/to/project
```

### Install with Go

```bash
go install github.com/funkykay/summarize/cmd/summarize@latest
```

Make sure the Go binary directory, usually `$(go env GOPATH)/bin`, is in your `PATH`.

### Standalone binaries

The `update` command expects GitHub release assets with these names:

| Platform | Asset name |
|---|---|
| Linux amd64 | `summarize-linux-x64` |
| Linux arm64 | `summarize-linux-arm64` |
| macOS arm64 | `summarize-macos-arm64` |

Windows and macOS Intel are not mapped by the current updater.

## Usage

```bash
summarize
summarize --base-dir path/to/project
summarize --profile minimal
summarize --dry-run
summarize --dry-run --profile minimal
summarize -p minimal --base-dir path/to/project
summarize init
summarize init --base-dir path/to/project
summarize update
summarize update --repo owner/repo
summarize version
```

Global options such as `--base-dir`, `--profile`, and `--dry-run` are parsed before the command name. The `init` command also accepts its own `--base-dir` after `init`.

## Output format

```text
=== relative/path/to/file ===
<file content>

=== another/file ===
<file content>
```

Paths are written relative to the selected base directory and normalized with forward slashes.

With `--dry-run`, `summarize` uses the same traversal, profile, prune, include, and exclude logic, but prints only the relative paths of files that would be exported:

```text
relative/path/to/file
another/file
```

Binary files are not printed as raw bytes. If a file is not valid UTF-8, the output contains:

```text
[Binary file - content not displayable]
```

If a directory cannot be read because of missing permissions, traversal continues and the output contains:

```text
[Access denied: relative/path]
```

## Configuration

A minimal `summarize.json`:

```json
{
  "selection_mode": "include_all",
  "include": [],
  "exclude": [],
  "prune": [
    ".git/"
  ]
}
```

Example with profiles:

```json
{
  "selection_mode": "include_all",
  "exclude": ["*.log", "dist/"],
  "prune": [".git/", "node_modules/"],
  "profiles": {
    "ci": {
      "selection_mode": "exclude_all",
      "include": ["cmd/", "internal/", "test/", "go.mod"],
      "exclude": ["internal/generated/"],
      "prune": ["coverage/"]
    }
  }
}
```

More details are in [`docs/configuration.md`](docs/configuration.md).

## Development

```bash
go test ./...
go build -o bin/summarize ./cmd/summarize
```

The project targets Go 1.22 and currently has no external Go module dependencies.

## Documentation

Start with [`docs/index.md`](docs/index.md) for the full technical documentation.


