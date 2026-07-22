package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Read decodes a JSON file and includes its path in parse errors.
func Read(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// WriteAtomic writes indented, newline-terminated JSON through a temporary file
// in the destination directory. JSON files managed by ThemeTime are private.
func WriteAtomic(path string, value any) (err error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if err = tmp.Chmod(0o600); err != nil {
		return err
	}
	if _, err = tmp.Write(data); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}
	return nil
}
