# Update mechanism

The `update` command is for standalone binary installations.

It is not involved in normal directory summarization and does not read `summarize.json`.

## Command

```bash
summarize update
summarize update --repo owner/repo
```

## Repository resolution

The repository slug is selected in this order:

1. `--repo owner/repo`
2. `GITHUB_REPOSITORY` environment variable
3. Built-in default: `funkykay/summarize`

The slug is used to call:

```text
https://api.github.com/repos/<owner/repo>/releases/latest
```

## Release response

The updater expects the latest release response to contain:

- `tag_name`
- `assets[].name`
- `assets[].browser_download_url`

If `tag_name` is empty, the command fails.

If the expected platform asset is missing, the command fails.

## Version normalization and comparison

Versions are normalized by trimming whitespace and removing a leading `v`.

Examples:

| Input | Normalized |
|---|---|
| `v0.8.1` | `0.8.1` |
| `0.8.1` | `0.8.1` |

Only the release part before `-` or `+` is parsed for numeric comparison.

Examples:

| Input | Parsed parts |
|---|---|
| `0.8.1` | `[0, 8, 1]` |
| `0.8.1-beta.1` | `[0, 8, 1]` |
| `1.0+build` | `[1, 0]` |

The remote version is newer when one of its numeric parts is greater than the local version at the same position. Missing parts are treated as zero.

Invalid numeric formats return an error.

## Supported assets

| GOOS | GOARCH | Asset name |
|---|---|---|
| `linux` | `amd64` | `summarize-linux-x64` |
| `linux` | `arm64` | `summarize-linux-arm64` |
| `darwin` | `arm64` | `summarize-macos-arm64` |

Unsupported platforms return an error like:

```text
No standalone release asset available for platform <system>/<arch>.
```

## Download behavior

The updater:

1. Resolves the currently running executable path with `os.Executable()`.
2. Creates the executable directory if needed.
3. Downloads the asset to a temporary file in that directory.
4. Copies the HTTP response body to the temp file.
5. Adds executable bits to the downloaded file.
6. Renames the temp file to `summarize.new` next to the current executable.

The existing executable is not replaced automatically.

After a successful download, the command prints:

```text
Downloaded summarize <version> to <path>/summarize.new.
Run the following command to replace the current binary:
mv <path>/summarize.new <path>/summarize
```

Paths are shell-quoted when needed.

## Failure behavior

The updater returns structured user-facing errors for common failure cases:

- Latest release cannot be fetched.
- GitHub returns a non-2xx status.
- Release response is not valid JSON.
- Release has no `tag_name`.
- Version format is unsupported.
- No platform asset exists.
- Asset download fails.
- Downloaded file cannot be written, chmodded, or renamed.

## Design rationale

The updater avoids replacing the running binary automatically. This keeps filesystem permissions, replacement timing, and rollback decisions under user control.
