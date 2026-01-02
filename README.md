# Directory Summarizer

A small, dependency-free CLI tool that recursively traverses a directory tree
and prints the contents of files in a stable, machine-readable stream.

Traversal and filtering are controlled by layered `summarize.json` configuration
files using gitignore-style (fnmatch-based) exclusion patterns.

This tool is designed for:
- feeding directory contents into LLMs or other tools,
- reproducible, text-based snapshots of codebases.

---

## Features

- Recursively walks a directory tree starting at the working directory
- Outputs each file preceded by a stable header:
  ```
  === relative/path/to/file ===
  ```
- Prints UTF-8 file contents verbatim
- Gracefully handles binary files and read errors
- Layered configuration via `summarize.json` files at arbitrary depths
- Gitignore-style exclusion patterns (documented below)
- Standard-library only (Python 3.10+)

---

## Installation

The tool is a single Python file. No installation step is required.

Recommended usage is via symlink:

```bash
ln -s /path/to/summarize.py summarize
```

Ensure it is executable:

```bash
chmod +x summarize
```

---

## Usage

Run from the directory you want to summarize:

```bash
summarize
```

The output is a stream of:

1. File marker:
   ```
   === relative/path ===
   ```
2. File contents (UTF-8)
3. A blank line separator

### Error Handling

- Binary or non-UTF-8 files:
  ```
  [Binary file - content not displayable]
  ```
- Unreadable files:
  ```
  [Error reading file: ...]
  ```
- Inaccessible directories:
  ```
  [Access denied: relative/path]
  ```

The tool never aborts the traversal due to a single file or directory.

---

## Configuration: `summarize.json`

The summarizer looks for a file named `summarize.json` in **every directory it enters**.

If found, that file is loaded as a new *configuration layer* that applies to:
- the directory itself, and
- all of its descendants.

When leaving the directory, the layer is automatically removed.

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

---

## Layering Semantics

Configuration layers are merged *top-down* using the following rules:

- `dict + dict` → recursive merge
- `list + list` → concatenation (cumulative)
- scalars → later layers overwrite earlier ones

This means:

- `excludes` are **additive** down the directory tree
- a deeper `summarize.json` does **not** replace parent excludes
- negation patterns (`!`) only affect paths matched later

### Example

```
.
├─ summarize.json        (excludes ".git/")
└─ project/
   ├─ summarize.json     (excludes "node_modules/")
   └─ src/
```

Running from `.` will exclude:
- `.git/`
- `project/node_modules/`

---

## Exclude Pattern Semantics

Patterns are **gitignore-like**, implemented using Python’s `fnmatch`.

Important: this is *not* a byte-for-byte reimplementation of Git’s ignore engine.
The behavior is deterministic and documented here.

### Supported Behavior

- Empty lines and lines starting with `#` are ignored
- Trailing `/` → directory-only match
  ```
  __pycache__/
  ```
- Leading `/` → path is anchored to the root (start directory)
  ```
  /build/output.txt
  ```
- Patterns without leading `/` are matched against:
  - the full relative path
  - each subpath segment
  - the filename / directory name itself
- Negation patterns (`!`) re-include matches (last match wins)

### Practical Implications

- `node_modules` will match **any** `node_modules` directory
- `*.log` matches logs at any depth
- `!/important.log` can re-include a previously excluded file

---

## Non-Goals

This tool intentionally does **not**:

- fully replicate Git’s ignore engine
- parse `.gitignore` files automatically
- follow symbolic links
- attempt parallel traversal
- redact secrets or sensitive data

It assumes you explicitly control what is included via configuration.

---

## Requirements

- Python 3.10 or newer
- UTF-8 locale recommended
