package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func ClearHostFolder(path string) error {
	if path == "" {
		return fmt.Errorf("Путь пустой")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if abs == "/" || abs == "." {
		return fmt.Errorf("Попытка удаления корневой папки прервана: %s", abs)
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil
	}

	err = os.RemoveAll(abs)
	if err != nil {
		return err
	}

	fmt.Println("Файлы успешно удалены из папки", path)
	return nil
}
