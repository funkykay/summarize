# Directory Summarizer

A small CLI utility that recursively traverses a directory tree and prints the contents of files.  
Traversal is controlled by layered `summarize.json` configuration files that provide gitignore-style (fnmatch-based) exclusion patterns.

## Features

- Recursively walks the current working directory
- Prints each file preceded by a header line: `=== relative/path/to/file ===`
- Skips files/directories using gitignore-style patterns
- Supports layered configuration: each directory can add/override config via `summarize.json`
- Configuration merges “down the tree”:
  - `dict + dict` merges recursively
  - `list + list` concatenates (e.g., excludes are cumulative)
  - scalars are overwritten by later layers

## Requirements

- Python 3.10+ (recommended)
- No external dependencies (standard library only)

## Installation (recommended)

Just link to some directory that is exposed to the path

```bash
ln -s /path/to/projects/summarize.py summarize
```


## Usage

Run from the directory you want to summarize:

```bash
./summarize.py
```
or
```bash
summarize
```

The tool will print a stream of:

- file marker: `=== relative/path ===`
- file contents (UTF-8)
- a blank line separator

If a file is not valid UTF-8, it prints:

```
[Binary file - content not displayable]
```

If a file cannot be read:

```
[Error reading file: ...]
```

## Configuration: `summarize.json`

The summarizer looks for a `summarize.json` in each directory it enters.  
That config is pushed as a new overlay layer while processing that directory and its descendants.

### Example

```json
{
  "excludes": [
    "__pycache__/",
    ".git/",
    "*.log",
    "node_modules/",
    "!node_modules/keep_me.txt"
  ]
}
```

### `excludes` patterns

Patterns are **gitignore-like** but implemented using Python `fnmatch`, so they are best understood as “gitignore-style matching,” not a byte-for-byte `.gitignore` equivalent.

Supported behavior:

- Empty lines and lines starting with `#` are ignored
- Trailing `/` indicates “directory only” (e.g., `__pycache__/`)
- Leading `/` anchors the match to the repository root (the directory where you ran the tool)
- Patterns without a leading `/` are matched against:
  - the full relative path, and
  - each subpath segment (so `node_modules` will match any `node_modules` directory)
- Negation patterns starting with `!` re-include matches (last match wins)

Important note: because config layers concatenate lists, `excludes` are **cumulative** down the tree. A deeper `summarize.json` adds additional excludes; it does not replace the parent excludes.

## How layering works (example)

Given:

- `./summarize.json` excludes `.git/`
- `./project/summarize.json` excludes `node_modules/`

Running from `.` will exclude both `.git/` and any `node_modules/` under `project/`.
