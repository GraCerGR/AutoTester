package commandonhost

import (
	"fmt"
	"os"
	"path/filepath"
)

func ClearHostFolder(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Защита от удаления корня
	if abs == "/" || abs == "." {
		return fmt.Errorf("refusing to clean dangerous path: %s", abs)
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil // папки нет — это нормально
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}

	for _, e := range entries {
		err := os.RemoveAll(filepath.Join(abs, e.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
