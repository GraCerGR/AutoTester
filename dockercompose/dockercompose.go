package dockercompose

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	hubStatusURL        = "http://localhost:4444/status"
	hubURL              = "http://localhost:4444"
	hubWaitTimeout      = 40 * time.Second
	hubWaitPollInterval = 1 * time.Second
)

func StartCompose(composeDir string) error {
	absDir, err := filepath.Abs(composeDir)
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь: %w", err)
	}

	fmt.Println("=== Запуск контейнеров через docker compose ===")

	if err := runCmd("docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "up", "-d"); err != nil {
		return fmt.Errorf("docker compose up failed: %w", err)
	}

	fmt.Println("=== Ожидание готовности Selenium Hub ...")
	if err := WaitForHub(hubWaitTimeout); err != nil {
		fmt.Println("=== Hub не стал готов вовремя, показываем логи:")
		_ = runCmd("docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "logs")
		return fmt.Errorf("hub not ready: %w", err)
	}

	fmt.Println("Selenium Grid запущен и готов по адресу:", hubURL)
	return nil
}

func StopCompose(composeDir string) error {
	absDir, err := filepath.Abs(composeDir)
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь: %w", err)
	}

	fmt.Println("=== Остановка Selenium Grid ===")
	if err := runCmd("docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "down", "-v"); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	fmt.Println("Selenium Grid остановлен")
	return nil
}

func runCmd(name string, args ...string) error {
	fmt.Printf(">>> running: %s %v\n", name, args)

	cmd := exec.Command(name, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command failed: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// WaitForHub опрашивает http://localhost:4444/status до таймаута.
func WaitForHub(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(hubStatusURL)
		if err == nil && resp.StatusCode == 200 {
			// успех
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(hubWaitPollInterval)
	}
	return fmt.Errorf("timeout waiting for hub at %s", hubStatusURL)
}
