#!/usr/bin/env python3
"""
Directory Summarizer - Recursively lists and displays file contents.

This tool traverses directories recursively and outputs the content of all files,
respecting gitignore-style exclusion patterns defined in summarize.json configuration files.
"""

from __future__ import annotations

import fnmatch
from pathlib import Path
from typing import Iterable

from config import StrataConfig


def gitignore_match(pattern: str, path_str: str, *, is_dir: bool) -> bool:
    """
    Check if a path matches a gitignore-style pattern.

    Args:
        pattern: The gitignore pattern to match against.
        path_str: The path as string (Unix-style with forward slashes).
        is_dir: True if the path represents a directory.

    Returns:
        True if the pattern matches, False otherwise.
    """
    p = pattern.strip()
    if not p or p.startswith("#"):
        return False

    # Pattern ending with / matches only directories
    if p.endswith("/"):
        if not is_dir:
            return False
        p = p[:-1]

    # Pattern starting with / matches only from root
    if p.startswith("/"):
        anchored = p[1:]
        return fnmatch.fnmatch(path_str, anchored)

    # Pattern without leading / matches anywhere in hierarchy
    parts = path_str.split("/")

    # Check against complete path
    if fnmatch.fnmatch(path_str, p):
        return True

    # Check against each subpath segment
    for i in range(len(parts)):
        subpath = "/".join(parts[i:])
        if fnmatch.fnmatch(subpath, p):
            return True
        if fnmatch.fnmatch(parts[i], p):
            return True

    return False


def is_excluded(path: Path, base_path: Path, exclude_patterns: Iterable[str]) -> bool:
    """
    Check if a path should be excluded based on gitignore-style patterns.

    Args:
        path: The path to check.
        base_path: The base path for relative path calculation.
        exclude_patterns: Iterable of gitignore-style patterns.

    Returns:
        True if the path should be excluded, False otherwise.
    """
    patterns = list(exclude_patterns)
    if not patterns:
        return False

    try:
        relative_path = path.relative_to(base_path)
    except ValueError:
        relative_path = path

    path_str = str(relative_path).replace("\\", "/")
    is_dir = path.is_dir()

    excluded = False
    for raw in patterns:
        p = raw.strip()
        if not p or p.startswith("#"):
            continue

        # Negation pattern (!) re-includes the path
        if p.startswith("!"):
            if gitignore_match(p[1:], path_str, is_dir=is_dir):
                excluded = False
            continue

        if gitignore_match(p, path_str, is_dir=is_dir):
            excluded = True

    return excluded


def process_directory_recursive(directory: Path, config: StrataConfig, base_path: Path) -> None:
    """
    Recursively process files and directories, outputting file contents.

    Args:
        directory: The directory to process.
        config: StrataConfig instance for layered configuration.
        base_path: The base path for relative path calculation.
    """
    layer_pushed = False
    try:
        config_file = directory / "summarize.json"
        layer_id = config.push_layer(config_file)
        layer_pushed = layer_id is not None

        exclude_patterns = config.get("excludes", [])

        for item in sorted(directory.iterdir()):
            if is_excluded(item, base_path, exclude_patterns):
                continue

            if item.is_dir():
                process_directory_recursive(item, config, base_path)
                continue

            relative_path = item.relative_to(base_path)
            print(f"=== {relative_path} ===")

            try:
                content = item.read_text(encoding="utf-8")
                print(content)
            except UnicodeDecodeError:
                print("[Binary file - content not displayable]")
            except Exception as exc:  # noqa: BLE001 - this is a CLI output tool
                print(f"[Error reading file: {exc}]")

            print()

    except PermissionError:
        relative_path = directory.relative_to(base_path)
        print(f"[Access denied: {relative_path}]")

    finally:
        if layer_pushed:
            config.pop_layer()


def main() -> int:
    """Main entry point for the directory summarizer."""
    base_dir = Path.cwd().resolve()
    config = StrataConfig()
    process_directory_recursive(base_dir, config, base_dir)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
