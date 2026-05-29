# CLI usage

## Command summary

```bash
summarize [--base-dir <dir>] [-p|--profile <name>] [--dry-run]
summarize init [--base-dir <dir>]
summarize update [--repo <owner/repo>]
summarize version
```

The default command is the summary command. It runs when no explicit subcommand is provided.

## Global options

| Option | Applies to | Description |
|---|---|---|
| `--base-dir <dir>` | default summary, also used as default for `init` when placed before `init` | Directory used as traversal root. Defaults to the current working directory. Must exist and be a directory. |
| `--base-dir=<dir>` | same | Inline form of `--base-dir`. |
| `-p <name>` | default summary | Selects a profile from `summarize.json`. |
| `--profile <name>` | default summary | Selects a profile from `summarize.json`. |
| `--profile=<name>` | default summary | Inline form of `--profile`. |
| `--dry-run` | default summary | Prints only the relative paths of files that would be exported. File contents and file-block headers are not printed. |

Global options are parsed before the first non-option argument. After the command name, options are command-specific.

Examples:

```bash
summarize --base-dir ./project --profile ci
summarize --base-dir ./project --profile ci --dry-run
summarize --base-dir ./project init
summarize init --base-dir ./project
```

## Default summary command

```bash
summarize
summarize --base-dir ./project
summarize --profile ci
summarize --dry-run
```

Execution steps:

1. Resolve `base_dir` to an absolute existing directory.
2. Load `base_dir/summarize.json` as the base configuration. A missing file is treated as an empty configuration.
3. Apply the selected profile if `--profile` or `-p` is present and the profile exists.
4. Traverse the directory tree.
5. Print included files to stdout. In dry-run mode, print only the included file paths.

### Dry run

```bash
summarize --dry-run
summarize --base-dir ./project --profile ci --dry-run
```

Dry run uses the same traversal, nested configuration, profile, prune, include, and exclude logic as a normal summary run. Instead of file blocks, stdout contains one normalized relative file path per line:

```text
README.md
docs/cli.md
internal/summary/summary.go
```

Dry-run output does not include `=== path ===` headers, file contents, binary placeholders, or access-denied markers. It is intended for checking which files would be exported before producing the full summary.

## `init`

```bash
summarize init
summarize init --base-dir ./project
```

Creates `summarize.json` in the target directory.

The command fails if `summarize.json` already exists.

The generated configuration uses:

```json
{
  "selection_mode": "include_all",
  "include": [],
  "exclude": [],
  "prune": []
}
```

Additional prune entries are added when project markers are detected:

| Detected marker | Added prune entry |
|---|---|
| `node_modules/` directory | `node_modules/` |
| `package.json` file | `node_modules/` |
| `.git/` directory | `.git/` |
| `.gitignore` file | `.git/` |

Entries are deduplicated while preserving rule order. The current rule order is `node_modules/` first, `.git/` second.

## `update`

```bash
summarize update
summarize update --repo funkykay/summarize
```

Checks the latest GitHub Release, compares it with the local version, downloads a matching standalone binary if the remote release is newer, and prints the `mv` command needed to replace the current executable manually.

The command does not modify the currently running binary automatically.

More details are in [Update mechanism](update.md).

## `version`

```bash
summarize version
```

Prints the value of `internal/buildinfo.Version` to stdout.

The current source version is:

```text
0.8.2
```

The command rejects additional arguments.

## Exit behavior

| Situation | Exit code |
|---|---|
| Successful command | `0` |
| CLI parse error | `1` |
| Runtime error | `1` |
| Unknown command | `1` |
| Invalid or missing option value | `1` |

Errors are written to stderr with the prefix `Error:`. Normal command output is written to stdout.

## Output details

In dry-run mode, every included file is printed as a single path line:

```text
path/from/base
```

For every included UTF-8 file in normal summary mode:

```text
=== path/from/base ===
<file content>
```

If a file is not valid UTF-8:

```text
=== path/from/base ===
[Binary file - content not displayable]
```

If a directory cannot be read because of missing permissions:

```text
[Access denied: path/from/base]
```

Traversal continues after access-denied directories.


