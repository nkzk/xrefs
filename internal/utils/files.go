package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func createFile(path string) error {
	_, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return nil
}

// Opens a file and creates it if it does not exit
// Remember to call defer file.Close() after calling this function
func OpenFile(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create directories for path %s: %w", path, err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := createFile(path)
			if err != nil {
				return nil, fmt.Errorf("file did not exist, failed to create: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to open file %s: %w", path, err)
		}
	}

	return file, nil
}
