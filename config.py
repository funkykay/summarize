"""
Layered JSON Configuration System

This module provides a hierarchical configuration system that supports:
- Base configuration with multiple overlay layers
- Gitignore-style pattern matching
- Deep merging of dictionaries and concatenation of lists
- File-based configuration loading with automatic fallbacks
"""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Optional, Union

JsonDict = dict[str, Any]
PathLike = Union[str, Path]


class StrataConfig:
    """
    Layered JSON configuration system supporting base config with overlay layers.

    Merge semantics:
    - dict + dict: recursive merge (extend)
    - list + list: concatenate (extend): base_list + overlay_list
    - other types: overlay overwrites base
    """

    def __init__(self, base: Optional[JsonDict] = None) -> None:
        self._base: JsonDict = dict(base or {})
        self._layers: list[_Layer] = []

    @classmethod
    def from_file(cls, path: PathLike, *, encoding: str = "utf-8") -> "StrataConfig":
        base = cls._load_json(path, encoding=encoding)
        return cls(base=base)

    def load_base(self, path: PathLike, *, encoding: str = "utf-8") -> None:
        """Load/replace the base configuration from file (missing file results in empty dict)."""
        self._base = self._load_json(path, encoding=encoding)

    def push_layer(
        self,
        path: PathLike,
        *,
        name: Optional[str] = None,
        encoding: str = "utf-8",
    ) -> Optional[str]:
        """
        Load a JSON file as a new overlay layer.

        Returns:
        - Layer ID (str) if a layer was created
        - None if the file is missing or empty (no layer created)
        """
        data = self._load_json(path, encoding=encoding)
        if not data:
            return None

        layer = _Layer(
            layer_id=_Layer.new_id(),
            name=name or Path(path).name,
            path=str(path),
            data=data,
        )
        self._layers.append(layer)
        return layer.layer_id

    def pop_layer(self) -> str:
        """Remove the most recently added layer (stack behavior) and return its ID."""
        if not self._layers:
            raise IndexError("No layer available to remove.")
        return self._layers.pop().layer_id

    def remove_layer(self, layer_id: str) -> None:
        """Remove a layer by its ID."""
        for i, layer in enumerate(self._layers):
            if layer.layer_id == layer_id:
                del self._layers[i]
                return
        raise KeyError(f"Layer with ID '{layer_id}' not found.")

    def clear_layers(self) -> None:
        """Remove all overlay layers, keeping the base configuration."""
        self._layers.clear()

    def list_layers(self) -> list[dict[str, str]]:
        """Return metadata for all active layers (from oldest to newest)."""
        return [{"id": l.layer_id, "name": l.name, "path": l.path} for l in self._layers]

    def get(self, dotted_path: str, default: Any = None) -> Any:
        """Get a value using dot notation. Returns default if not found."""
        try:
            return self._get_or_raise(dotted_path)
        except KeyError:
            return default

    def require(self, dotted_path: str) -> Any:
        """Like get(), but raises KeyError if the path doesn't exist."""
        return self._get_or_raise(dotted_path)

    def has(self, dotted_path: str) -> bool:
        """Return True if the path exists."""
        try:
            self._get_or_raise(dotted_path)
        except KeyError:
            return False
        return True

    def to_dict(self) -> JsonDict:
        """Return the effectively merged configuration as a new dictionary."""
        merged: JsonDict = dict(self._base)
        for layer in self._layers:
            merged = self._deep_merge(merged, layer.data)
        return merged

    def __getitem__(self, dotted_path: str) -> Any:
        return self.require(dotted_path)

    def _get_or_raise(self, dotted_path: str) -> Any:
        keys = self._split_path(dotted_path)
        node: Any = self.to_dict()

        for depth, key in enumerate(keys):
            if not isinstance(node, dict):
                raise KeyError(
                    f"Path not found: '{dotted_path}' "
                    f"(intermediate node is not an object/dict at level {depth})"
                )
            if key not in node:
                raise KeyError(
                    f"Path not found: '{dotted_path}' (missing key '{key}' at level {depth})"
                )
            node = node[key]

        return node

    @staticmethod
    def _split_path(dotted_path: str) -> list[str]:
        p = dotted_path.strip()
        if not p:
            raise ValueError("dotted_path cannot be empty.")
        parts = [x for x in p.split(".") if x]
        if not parts:
            raise ValueError("Invalid dotted_path.")
        return parts

    @staticmethod
    def _load_json(path: PathLike, *, encoding: str = "utf-8") -> JsonDict:
        """
        Load JSON from file.
        - Missing file => {}
        - Root must be dict, otherwise ValueError
        """
        p = Path(path)
        if not p.exists():
            return {}

        with p.open("r", encoding=encoding) as f:
            data = json.load(f)

        if not isinstance(data, dict):
            raise ValueError(f"JSON root must be an object (dict), got: {type(data).__name__}")

        return data

    @classmethod
    def _deep_merge(cls, base: Any, overlay: Any) -> Any:
        """
        Merge rules:
        - dict+dict: recursive merge (combine keys)
        - list+list: concatenate
        - other: overlay wins
        """
        if isinstance(base, dict) and isinstance(overlay, dict):
            result: JsonDict = dict(base)
            for k, v in overlay.items():
                if k in result:
                    result[k] = cls._deep_merge(result[k], v)
                else:
                    result[k] = v
            return result

        if isinstance(base, list) and isinstance(overlay, list):
            return list(base) + list(overlay)

        return overlay


@dataclass(frozen=True)
class _Layer:
    layer_id: str
    name: str
    path: str
    data: JsonDict

    @staticmethod
    def new_id() -> str:
        _LayerIdCounter.next()
        return f"layer-{_LayerIdCounter.value:06d}"


class _LayerIdCounter:
    value: int = 0

    @classmethod
    def next(cls) -> None:
        cls.value += 1
