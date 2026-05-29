# Risks and technical debt

## Known risks

### Update command has no dedicated tests in the current export

The updater performs HTTP requests, platform selection, filesystem writes, chmod, and rename operations. The current project export does not include dedicated tests for this package.

Recommended test additions:

- Mock `HTTPClient` responses for latest release metadata.
- Test non-2xx GitHub responses.
- Test invalid JSON responses.
- Test asset selection failure.
- Test version comparison edge cases.
- Test successful binary download into a temp directory.

### Version comparison is intentionally simple

The updater parses numeric release parts and strips suffixes after `-` or `+`. This is sufficient for stable numeric releases such as `v0.8.1`, but it is not a complete SemVer implementation.

Pre-release ordering is not modeled.

### Missing config validation beyond JSON root type

The config loader validates that JSON is syntactically valid and that the root is an object. Most schema errors are silently ignored by downstream normalization or result in default behavior.

Examples:

- `include` as a string instead of an array is ignored.
- Invalid profile `selection_mode` is ignored.
- Unknown keys are retained but unused.

This makes the tool permissive, but it can hide user mistakes.

### Unknown profiles are silent

Selecting a missing profile does not warn or fail. This avoids breaking scripts but can make typos hard to notice.

### Anchored patterns are anchored to base directory

Nested `summarize.json` files can define anchored patterns, but anchoring is still relative to the original traversal base directory. This may be surprising if a user expects `/file.txt` inside `subdir/summarize.json` to mean `subdir/file.txt`.

### Whole-file reads

File contents are read fully into memory before output. This keeps implementation simple but can be expensive for very large files.

### UTF-8-only text output

Files that are not valid UTF-8 are treated as binary and are not decoded with fallback encodings such as Latin-1 or Shift-JIS.

### Limited platform support in updater

The update command maps only Linux amd64, Linux arm64, and macOS arm64 to release assets. Windows and macOS Intel are unsupported by the updater.

### Global layer ID counter

Config layer IDs are generated from a package-level counter. This is simple and stable for normal CLI use, but it is not designed as a concurrency-safe API.

## Technical debt

### No `--output` option

The summary is written only to stdout. Users must redirect output manually:

```bash
summarize > summary.txt
```

A native `--output` option could provide clearer errors and safer overwrite behavior.

### No structured output format

The plain-text block format is intentionally simple. A JSON output mode could help consumers that need unambiguous machine parsing, but it would add compatibility and escaping concerns.

### Hand-written CLI parser

The parser is small and dependency-free, but every new command or option requires manual parsing logic and tests.

### Pattern matcher is custom

The gitignore-like matcher is implemented in the project. It supports the patterns currently needed by tests, but it is not a full Git implementation.

## Open decisions

- Should selecting an unknown profile produce a warning?
- Should missing `summarize.json` remain valid default behavior?
- Should anchored patterns in nested configs be relative to the nested config directory instead of the base directory?
- Should update support Windows and macOS Intel assets?
- Should release version parsing use a full SemVer implementation?
- Should large files have a size limit or streaming behavior?
