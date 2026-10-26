package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// LibraryPaths resolves named files and immediate directory entries. Directory
// entries use their basename without extension, across all source extensions.
// Duplicate names are errors, even if the conflicting files have equal content.
func (cfg Config) LibraryPaths(root string) (map[string]string, error) {
	paths := map[string]string{}
	add := func(name, path string, stat os.FileInfo) error {
		if _, exists := paths[name]; exists {
			return fmt.Errorf("duplicate library name %q (%s)", name, path)
		}
		if stat == nil {
			var err error
			stat, err = os.Stat(path)
			if err != nil {
				return fmt.Errorf("library %s: %w", name, err)
			}
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("library %s is not a regular file: %s", name, path)
		}
		paths[name] = path
		return nil
	}
	for _, name := range slices.Sorted(maps.Keys(cfg.Lib.Include)) {
		path := ResolvePath(root, cfg.Lib.Include[name])
		stat, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("library %s: %w", name, err)
		}
		if !stat.IsDir() {
			if err := add(name, path, stat); err != nil {
				return nil, err
			}
			continue
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("library directory %s: %w", path, err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			if err := add(key, filepath.Join(path, entry.Name()), nil); err != nil {
				return nil, err
			}
		}
	}
	return paths, nil
}
