# Data flow

## Default summary flow

```text
1. User runs `summarize --base-dir /project --profile ci`.

2. `cmd/summarize/main.go` passes args and stdio writers to `cli.App.Run`.

3. `internal/cli` parses global options.
   - `--base-dir` is resolved to an absolute existing directory.
   - `--profile` is stored as a string.
   - No command name means default summary mode.

4. `internal/summary.Create` loads `/project/summarize.json`.
   - Missing file becomes an empty config.
   - Invalid JSON or non-object JSON returns an error.

5. `internal/profile.ApplyLayer` applies the selected profile.
   - Missing profile: no change.
   - Valid profile: push `profile:<name>` layer.

6. `internal/summary.Directory` starts traversal at `/project`.
   - The root config is already loaded, so the root directory is not loaded again as a nested layer.

7. For every directory:
   - Optionally push local `summarize.json` as a layer.
   - Read entries.
   - Sort entries by name.
   - Evaluate prune rules for every entry.
   - Recurse into non-pruned directories.
   - Evaluate selection rules for non-pruned files.
   - Print selected file blocks.
   - Pop the local config layer before returning.

8. `cli.App.Run` returns `0` on success or `1` on error.
```

## Traversal decision order

For each filesystem entry:

```text
entry path
  -> prune check
     -> pruned: skip entry completely
     -> not pruned:
        -> directory: recurse
        -> file: selection check
           -> allowed: print file block
           -> rejected: skip file
```

Prune runs before selection. This is important because pruning can avoid reading entire directory subtrees.

## Configuration layer timeline

Example tree:

```text
project/
  summarize.json
  src/
    summarize.json
    app.go
  test/
    helper.go
```

Execution timeline:

```text
load project/summarize.json as base
apply optional profile layer
enter project/              no push; root config is already base
enter project/src/          push project/src/summarize.json if non-empty
process project/src/app.go  use base + profile + src layer
leave project/src/          pop src layer
enter project/test/         no local layer
process project/test/helper.go use base + profile only
```

## Merge timing

The effective configuration is calculated when `ToMap`, `Get`, or `Require` is called. It is not permanently flattened after every layer change.

This keeps push/pop traversal reversible:

```text
base + profile + src layer
pop src layer
base + profile
```

## Summary output generation

For a selected file:

1. Calculate the path relative to the traversal base directory.
2. Convert path separators to `/`.
3. Print the header:

   ```text
   === relative/path ===
   ```

4. Read the whole file into memory.
5. If bytes are valid UTF-8, print the content.
6. If bytes are not valid UTF-8, print the binary placeholder.
7. Print a blank line after the file block.

## Permission errors

If reading a directory fails with `os.ErrPermission`, the error is converted into output:

```text
[Access denied: relative/path]
```

Traversal then continues. Other directory read errors are returned as command errors.

## `init` flow

```text
1. Parse optional command-specific `--base-dir`.
2. Resolve the directory and require it to exist.
3. Check whether `summarize.json` already exists.
4. Build initial config from rules.
5. JSON-encode with two-space indentation and trailing newline.
6. Write `summarize.json` with mode `0644`.
7. Print `Created <path>`.
```

## `update` flow

```text
1. Parse optional `--repo`.
2. Determine repository slug.
3. Fetch latest GitHub Release metadata.
4. Normalize local and remote versions.
5. Compare numeric release parts.
6. Stop if local version is already up to date.
7. Determine platform-specific asset name.
8. Find the asset download URL.
9. Detect the current executable path.
10. Download to a temp file in the executable directory.
11. Mark the temp file executable.
12. Rename temp file to `summarize.new`.
13. Print the manual replacement command.
```
