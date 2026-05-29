# Glossary

## Base directory

The root directory for traversal. Defaults to the current working directory and can be changed with `--base-dir`.

All printed file paths are relative to the base directory.

## Summary

The generated plain-text output containing one block per included file.

## Dry run

A preview mode activated with `--dry-run`. It uses the normal traversal and selection logic, but prints only the relative paths of files that would be exported.

## File block

One output section in this format:

```text
=== relative/path ===
<file content>
```

## Prune

A pre-traversal exclusion decision. If a directory is pruned, the whole subtree is skipped. If a file is pruned, it is not considered for selection or output.

Prune is configured through the `prune` array.

## Selection

The file-level decision that determines whether a non-pruned file is printed.

Selection is configured through `selection_mode`, `include`, and `exclude`.

## Selection mode

The default posture of file selection.

- `include_all`: include files unless excluded; include can override exclude.
- `exclude_all`: exclude files unless included; exclude overrides include.

## Include pattern

A pattern in the `include` array. It explicitly includes matching files according to the active selection mode.

## Exclude pattern

A pattern in the `exclude` array. It explicitly excludes matching files according to the active selection mode.

## Prune pattern

A pattern in the `prune` array. It can skip files or directories before selection. Prune patterns support `!` negation.

## Negation pattern

A prune pattern beginning with `!`, for example `!important.log`. It cancels an earlier prune match when it matches later in the prune list.

Negation is only handled by prune evaluation, not by include or exclude selection.

## Anchored pattern

A pattern beginning with `/`, for example `/go.mod`. It matches relative to the traversal base directory.

## Directory pattern

A pattern ending with `/`, for example `build/`. It matches directories with that name. In selection, descendant-aware matching lets it affect files below the directory.

## Layer

One configuration overlay inside `internal/config.Config`. Layers can come from profiles or nested `summarize.json` files.

## Base config

The configuration loaded from `<base-dir>/summarize.json`. If the file is missing, the base config is empty.

## Profile

A named configuration overlay under the `profiles` object in `summarize.json`. Profiles are activated with `--profile` or `-p`.

## Nested config

A `summarize.json` file inside a subdirectory. It is pushed as a temporary layer while that directory subtree is traversed.

## Deep merge

The config merge strategy:

- Objects merge recursively.
- Arrays concatenate.
- Scalars are replaced by overlay values.

## Standalone binary

An executable built for a specific OS and CPU architecture. The `update` command is designed for this installation style.

## Repository slug

A GitHub repository identifier in `owner/repo` format.

## Update candidate

The downloaded replacement binary written as `summarize.new`. It becomes active only if the user manually moves it over the current binary.


