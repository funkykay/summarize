package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Dict map[string]any

type Config struct {
	base   Dict
	layers []Layer
}

type Layer struct {
	ID   string
	Name string
	Path string
	Data Dict
}

type LayerInfo struct {
	ID   string
	Name string
	Path string
}

var layerIDCounter int

func New(base Dict) *Config {
	copied := Dict{}
	for key, value := range base {
		copied[key] = value
	}

	return &Config{base: copied}
}

func FromFile(path string) (*Config, error) {
	base, err := loadJSON(path)
	if err != nil {
		return nil, err
	}

	return New(base), nil
}

func (c *Config) LoadBase(path string) error {
	base, err := loadJSON(path)
	if err != nil {
		return err
	}

	c.base = base
	return nil
}

func (c *Config) PushLayer(path string) (string, bool, error) {
	data, err := loadJSON(path)
	if err != nil {
		return "", false, err
	}
	if len(data) == 0 {
		return "", false, nil
	}

	id := c.PushDataLayer(data, layerName(path), path)
	return id, true, nil
}

func (c *Config) PushDataLayer(data Dict, name string, path string) string {
	layerIDCounter++
	id := fmt.Sprintf("layer-%06d", layerIDCounter)

	copied := Dict{}
	for key, value := range data {
		copied[key] = value
	}

	c.layers = append(c.layers, Layer{
		ID:   id,
		Name: name,
		Path: path,
		Data: copied,
	})

	return id
}

func (c *Config) PopLayer() (string, error) {
	if len(c.layers) == 0 {
		return "", errors.New("no layer available to remove")
	}

	last := len(c.layers) - 1
	id := c.layers[last].ID
	c.layers = c.layers[:last]
	return id, nil
}

func (c *Config) RemoveLayer(layerID string) error {
	for i, layer := range c.layers {
		if layer.ID == layerID {
			c.layers = append(c.layers[:i], c.layers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("layer with ID %q not found", layerID)
}

func (c *Config) ClearLayers() {
	c.layers = nil
}

func (c *Config) ListLayers() []LayerInfo {
	infos := make([]LayerInfo, 0, len(c.layers))
	for _, layer := range c.layers {
		infos = append(infos, LayerInfo{
			ID:   layer.ID,
			Name: layer.Name,
			Path: layer.Path,
		})
	}

	return infos
}

func (c *Config) Get(dottedPath string, fallback any) any {
	value, err := c.Require(dottedPath)
	if err != nil {
		return fallback
	}

	return value
}

func (c *Config) Require(dottedPath string) (any, error) {
	keys, err := splitPath(dottedPath)
	if err != nil {
		return nil, err
	}

	var node any = c.ToMap()
	for depth, key := range keys {
		object, ok := node.(Dict)
		if !ok {
			if generic, genericOK := node.(map[string]any); genericOK {
				object = Dict(generic)
			} else {
				return nil, fmt.Errorf("path not found: %q (intermediate node is not an object/dict at level %d)", dottedPath, depth)
			}
		}

		value, exists := object[key]
		if !exists {
			return nil, fmt.Errorf("path not found: %q (missing key %q at level %d)", dottedPath, key, depth)
		}
		node = value
	}

	return node, nil
}

func (c *Config) Has(dottedPath string) bool {
	_, err := c.Require(dottedPath)
	return err == nil
}

func (c *Config) ToMap() Dict {
	merged := cloneDict(c.base)
	for _, layer := range c.layers {
		merged = deepMerge(merged, layer.Data).(Dict)
	}

	return merged
}

func loadJSON(path string) (Dict, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Dict{}, nil
		}
		return nil, err
	}

	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	object, ok := decoded.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("JSON root must be an object (dict), got: %s", jsonTypeName(decoded))
	}

	return Dict(object), nil
}

func splitPath(dottedPath string) ([]string, error) {
	trimmed := strings.TrimSpace(dottedPath)
	if trimmed == "" {
		return nil, errors.New("dotted_path cannot be empty")
	}

	rawParts := strings.Split(trimmed, ".")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if part != "" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return nil, errors.New("invalid dotted_path")
	}

	return parts, nil
}

func deepMerge(base any, overlay any) any {
	baseMap, baseIsMap := asDict(base)
	overlayMap, overlayIsMap := asDict(overlay)
	if baseIsMap && overlayIsMap {
		result := cloneDict(baseMap)
		for key, value := range overlayMap {
			if existing, exists := result[key]; exists {
				result[key] = deepMerge(existing, value)
			} else {
				result[key] = value
			}
		}
		return result
	}

	baseList, baseIsList := asList(base)
	overlayList, overlayIsList := asList(overlay)
	if baseIsList && overlayIsList {
		result := make([]any, 0, len(baseList)+len(overlayList))
		result = append(result, baseList...)
		result = append(result, overlayList...)
		return result
	}

	return overlay
}

func cloneDict(source Dict) Dict {
	result := Dict{}
	for key, value := range source {
		result[key] = value
	}

	return result
}

func asDict(value any) (Dict, bool) {
	switch typed := value.(type) {
	case Dict:
		return typed, true
	case map[string]any:
		return Dict(typed), true
	default:
		return nil, false
	}
}

func asList(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		result := make([]any, 0, len(typed))
		for _, entry := range typed {
			result = append(result, entry)
		}
		return result, true
	default:
		return nil, false
	}
}

func layerName(path string) string {
	name := filepath.Base(path)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return path
	}

	return name
}

func jsonTypeName(value any) string {
	switch value.(type) {
	case nil:
		return "NoneType"
	case bool:
		return "bool"
	case float64:
		return "float"
	case string:
		return "str"
	case []any:
		return "list"
	case map[string]any:
		return "dict"
	default:
		return fmt.Sprintf("%T", value)
	}
}
