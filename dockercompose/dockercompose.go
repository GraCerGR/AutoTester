package dockercompose

import (
	"MainApp/containermanager"
	"MainApp/settings"
	"context"
	"fmt"
	"path/filepath"
)

func StartCompose(ctx context.Context, composeDir string, profiles []string) error {
	absDir, err := filepath.Abs(composeDir)
	if err != nil {
		return fmt.Errorf("Не удалось получить абсолютный путь: %w", err)
	}

	fmt.Println("Запуск контейнеров docker compose")

	args := []string{
		"compose",
		"-f", filepath.Join(absDir, "docker-compose.yml"),
	}

	for _, p := range profiles {
		args = append(args, "--profile", p)
	}

	args = append(args, "up", "-d")

	if Contains(profiles, "selenium") {
		args = append(args, "--scale", "selenium-node-chrome="+settings.SeleniumNodeChromeNumber)
	}

	if err := containermanager.RunCmd(ctx, "docker", args...); err != nil {
		return fmt.Errorf("docker compose up failed: %w", err)
	}

	return nil
}

func StopCompose(ctx context.Context, composeDir string) error {
	absDir, err := filepath.Abs(composeDir)
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь: %w", err)
	}

	fmt.Println("=== Остановка Selenium Grid ===")
	if err := containermanager.RunCmd(ctx, "docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "down", "-v"); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	fmt.Println("Selenium Grid остановлен")
	return nil
}

func Contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}