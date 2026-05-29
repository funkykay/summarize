# Configuration

## `summarize.json`

`summarize.json` is a JSON object. The root must be an object; any other JSON root type is an error.

Supported keys used by the current implementation:

| Key | Type | Default behavior | Description |
|---|---|---|---|
| `selection_mode` | string | `include_all` | File selection mode. Valid values are `include_all` and `exclude_all`. |
| `include` | string array | empty | Patterns that explicitly include files. |
| `exclude` | string array | empty | Patterns that explicitly exclude files. |
| `prune` | string array | empty | Patterns that prune files or directories before traversal/output. |
| `profiles` | object | empty | Named overlays activated with `--profile` or `-p`. |

Unknown keys are preserved in the merged configuration map but are not used by traversal, selection, profile handling, initialization, or updating.

A missing `summarize.json` is treated as an empty configuration.

## Minimal configuration

```json
{
  "selection_mode": "include_all",
  "include": [],
  "exclude": [],
  "prune": []
}
```

## Example configuration

```json
{
  "selection_mode": "include_all",
  "include": ["README.md"],
  "exclude": ["*.log", "dist/"],
  "prune": [".git/", "node_modules/"],
  "profiles": {
    "minimal": {
      "selection_mode": "exclude_all",
      "include": ["cmd/", "internal/", "go.mod"],
      "exclude": ["internal/generated/"],
      "prune": ["coverage/"]
    }
  }
}
```

Run with the profile:

```bash
summarize --profile minimal
```

## Selection modes

### `include_all`

Default behavior.

A file is included unless it is excluded. `include` can re-include a path that also matches `exclude`.

Decision rule:

```text
included = explicitly_included OR NOT explicitly_excluded
```

Example:

```json
{
  "selection_mode": "include_all",
  "exclude": ["*.log"],
  "include": ["important.log"]
}
```

`important.log` is included even though it also matches `*.log`.

### `exclude_all`

A file is excluded unless it is explicitly included. `exclude` wins over `include`.

Decision rule:

```text
included = explicitly_included AND NOT explicitly_excluded
```

Example:

```json
{
  "selection_mode": "exclude_all",
  "include": ["src/"],
  "exclude": ["src/generated/"]
}
```

Files below `src/` are included except files below `src/generated/`.

Invalid `selection_mode` values are interpreted as `include_all` in normal traversal. In profile overlays, invalid `selection_mode` values are ignored so the previously active mode remains unchanged.

## Include and exclude patterns

`include` and `exclude` use the same gitignore-like matcher as prune patterns, but without `!` negation handling.

Directory patterns can match descendants during selection:

```json
{
  "selection_mode": "exclude_all",
  "include": ["src/"]
}
```

This includes files below `src/`, not only the directory entry itself.

## Prune patterns

`prune` is evaluated before directory recursion and before file output. It can skip whole subtrees and can also skip individual files.

Prune supports negation with `!`:

```json
{
  "prune": ["*.log", "!important.log"]
}
```

Patterns are evaluated from left to right. The last matching prune rule wins.

Example:

```json
{
  "prune": ["*.log", "!important.log", "debug.log"]
}
```

| Path | Result |
|---|---|
| `server.log` | pruned |
| `important.log` | not pruned |
| `debug.log` | pruned |

## Pattern behavior

| Pattern | Meaning |
|---|---|
| `*.log` | Matches log files at any level. |
| `build/` | Matches directories named `build`. In selection, descendants are affected. In pruning, the directory is skipped before recursion. |
| `/go.mod` | Matches `go.mod` only at the traversal root. |
| `docs/*.md` | Matches Markdown files directly below `docs`. |
| `!important.log` | Only meaningful in `prune`; cancels an earlier prune match. |

Matching internally uses forward slashes. Platform-specific path separators are normalized before matching.

Anchored patterns with a leading `/` are anchored to the original base directory, not to the directory containing a nested `summarize.json`.

## Layered configuration

Configuration is represented as a base map plus zero or more layers.

Layer order during a summary run:

1. Base configuration from `<base-dir>/summarize.json`.
2. Optional profile layer from `profiles.<name>`.
3. Nested directory layers from `summarize.json` files found while traversing subdirectories.

The root `summarize.json` is loaded once as the base configuration and is not pushed again when traversing the root directory.

Nested `summarize.json` files are pushed when entering their directory and popped after leaving it. Empty nested configuration files are ignored.

## Merge rules

When the effective configuration is requested, layers are merged bottom-up.

| Value type | Merge behavior |
|---|---|
| Object | Recursively merged by key. |
| Array | Concatenated: base array followed by overlay array. |
| Scalar | Overlay replaces base. |

Example:

Base:

```json
{
  "selection_mode": "include_all",
  "prune": [".git/"],
  "exclude": ["*.log"]
}
```

Profile:

```json
{
  "selection_mode": "exclude_all",
  "prune": ["coverage/"],
  "include": ["cmd/", "internal/"]
}
```

Effective configuration:

```json
{
  "selection_mode": "exclude_all",
  "prune": [".git/", "coverage/"],
  "exclude": ["*.log"],
  "include": ["cmd/", "internal/"]
}
```

## Profiles

Profiles live below `profiles`:

```json
{
  "profiles": {
    "ci": {
      "selection_mode": "exclude_all",
      "include": ["cmd/", "internal/", "test/"],
      "prune": ["coverage/"]
    }
  }
}
```

Supported profile keys:

- `selection_mode`
- `include`
- `exclude`
- `prune`

Profile `selection_mode` is applied only when it is exactly `include_all` or `exclude_all` after trimming and lowercasing. Invalid values are ignored.

Profile pattern arrays are normalized by trimming entries and dropping empty strings. Non-string entries are ignored.

Unknown profile names do not produce an error and do not change the output.


