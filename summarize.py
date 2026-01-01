#!/usr/bin/env python3
"""
Directory Summarizer - Recursively lists and displays file contents.

This tool traverses directories recursively and outputs the content of all files,
respecting gitignore-style exclusion patterns defined in summarize.json configuration files.
"""

import fnmatch
from pathlib import Path
from typing import List
from config import StrataConfig

def gitignore_match(pattern: str, path_str: str, is_dir: bool) -> bool:
    """
    Check if a path matches a gitignore-style pattern.
    
    Args:
        pattern: The gitignore pattern to match against
        path_str: The path as string (Unix-style with forward slashes)
        is_dir: True if the path represents a directory
        
    Returns:
        True if the pattern matches, False otherwise
    """
    # Skip empty patterns or comments
    if not pattern or pattern.startswith('#'):
        return False
    
    # Remove whitespace
    pattern = pattern.strip()
    if not pattern:
        return False
    
    # Pattern ending with / matches only directories
    if pattern.endswith('/'):
        if not is_dir:
            return False
        pattern = pattern[:-1]  # Remove trailing slash
    
    # Pattern starting with / matches only from root
    if pattern.startswith('/'):
        pattern = pattern[1:]  # Remove leading slash
        return fnmatch.fnmatch(path_str, pattern)
    
    # Pattern without leading / matches anywhere in hierarchy
    path_parts = path_str.split('/')
    
    # Check against complete path
    if fnmatch.fnmatch(path_str, pattern):
        return True
    
    # Check against each subpath (for patterns like "*.log" that should match everywhere)
    for i in range(len(path_parts)):
        subpath = '/'.join(path_parts[i:])
        if fnmatch.fnmatch(subpath, pattern):
            return True
        
        # Also check just the file/directory name
        if fnmatch.fnmatch(path_parts[i], pattern):
            return True
    
    return False

def is_excluded(path: Path, base_path: Path, exclude_patterns: List[str]) -> bool:
    """
    Check if a path should be excluded based on gitignore-style patterns.
    
    Args:
        path: The path to check
        base_path: The base path for relative path calculation
        exclude_patterns: List of gitignore-style patterns for exclusion
        
    Returns:
        True if the path should be excluded, False otherwise
    """
    if not exclude_patterns:
        return False
    
    # Calculate relative path
    try:
        relative_path = path.relative_to(base_path)
    except ValueError:
        # If path is not under base_path, use absolute path
        relative_path = path
    
    # Convert path to string with forward slashes (Unix-style)
    path_str = str(relative_path).replace('\\', '/')
    is_dir = path.is_dir()
    
    # Process all patterns using gitignore logic
    excluded = False
    
    for pattern in exclude_patterns:
        pattern = pattern.strip()
        if not pattern or pattern.startswith('#'):
            continue
            
        # Negation pattern (!) re-includes the path
        if pattern.startswith('!'):
            negation_pattern = pattern[1:]
            if gitignore_match(negation_pattern, path_str, is_dir):
                excluded = False
        else:
            # Normal exclusion pattern
            if gitignore_match(pattern, path_str, is_dir):
                excluded = True
    
    return excluded


def process_directory_recursive(directory: Path, config: StrataConfig, base_path: Path) -> None:
    """
    Recursively process all files and directories, outputting file contents.
    
    Args:
        directory: The directory to process
        config: StrataConfig object for configuration management
        base_path: The base path for relative path calculation
    """
    try:
        # Push config layer for current directory (if summarize.json exists)
        config_file = directory / "summarize.json"
        layer_id = config.push_layer(config_file)
        
        # Load exclusion patterns from config
        exclude_patterns = config.get("excludes", [])
        
        # Process all entries in the directory (sorted)
        for item in sorted(directory.iterdir()):
            # Check if item should be excluded
            if is_excluded(item, base_path, exclude_patterns):
                continue
            
            if item.is_dir():
                # Recursively process subdirectory
                process_directory_recursive(item, config, base_path)
            else:
                # Calculate relative path and output file
                relative_path = item.relative_to(base_path)
                print(f"=== {relative_path} ===")
                
                # Read and output file content
                try:
                    with open(item, 'r', encoding='utf-8') as f:
                        content = f.read()
                        print(content)
                except UnicodeDecodeError:
                    # Handle binary files that can't be read as UTF-8
                    print("[Binary file - content not displayable]")
                except Exception as e:
                    print(f"[Error reading file: {e}]")
                
                print()  # Empty line between files
        
        # Remove config layer if one was pushed
        if layer_id is not None:
            config.remove_layer(layer_id)
                
    except PermissionError:
        # Handle directories we don't have permission to read
        relative_path = directory.relative_to(base_path)
        print(f"[Access denied: {relative_path}]")

def main() -> int:
    """Main entry point for the directory summarizer."""
    base_dir = Path.cwd().resolve()
    
    # Initialize StrataConfig object
    config = StrataConfig()
    
    # Start recursive processing
    process_directory_recursive(base_dir, config, base_dir)
    
    return 0


if __name__ == "__main__":
    raise SystemExit(main())