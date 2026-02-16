package conteinermanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunPythonTestsContainer(containerName string) (bool, error) {
	if isRunning, err := checkContainerRunning(containerName); err != nil || !isRunning {
		return false, fmt.Errorf("Контейнер %s не запущен: %w", containerName, err)
	}

	args := []string{
		"exec",
		"-u", "seluser",
		"-w", "/app",
		"-e", "PYTHONPATH=/app",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		containerName,
		"pytest", "-q", "-v", "-s", "--junitxml=/app/results/results.xml",
	}

	fmt.Printf("Запуск Python тестов: docker %v\n", args)

	passed, err := runCmdAllowFail("docker", args...)
	if err != nil {
		return false, err
	}
	if passed {
		fmt.Println("Python тесты прошли")
	} else {
		fmt.Println("Python тесты не прошли (но продолжаем)")
	}

	return passed, nil
}

func SendSitePythonCustomize(dir, hostFile, containerName string) error {
	if hostFile == "" {
		return fmt.Errorf("hostFile must be provided")
	}
	if _, err := os.Stat(filepath.Join(dir, hostFile)); err != nil {
		return fmt.Errorf("файл не найден: %w", err)
	}

	// Копируем в /app контейнера
	dst := fmt.Sprintf("%s:/app/%s", containerName, hostFile)
	if err := RunCmd("docker", "cp", filepath.Join(dir, hostFile), dst); err != nil {
		return fmt.Errorf("ошибка копирования: %w", err)
	}

	fmt.Println(hostFile + " скопирован в /app контейнера")
	return nil
}

func ReplaceTestURLInPythonContainer(containerName, varName, newURL string) error {
	// Экранируем символы для sed (вставим обратный слэш перед & | \)
	escapedURL := strings.NewReplacer(
		`&`, `\&`,
		`|`, `\|`,
		`\`, `\\`,
	).Replace(newURL)

	// sed: заменить строку вида varName = "...".
	sedExpr := fmt.Sprintf(`s|^\(%s\s*=\s*\).*|\1"%s"|`, varName, escapedURL)

	cmd := exec.Command(
		"docker", "exec", containerName,
		"bash", "-c",
		fmt.Sprintf(`find /app -name '*.py' -exec sed -i '%s' {} +`, sedExpr),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func CopyResultsFromPythonContainer(container, hostPath string) error {
	os.MkdirAll(hostPath, 0755)
	return RunCmd("docker", "cp",
		container+":/app/results/results.xml",
		filepath.Join(hostPath, "results.xml"),
	)
}
