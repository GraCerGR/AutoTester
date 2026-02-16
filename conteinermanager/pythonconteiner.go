package conteinermanager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunPythonTestsContainer(image string) (bool, error) {
	checkCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", image))
	out, err := checkCmd.Output()
	if err != nil {
		return false, fmt.Errorf("не удалось выполнить проверку контейнеров: %w", err)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return false, fmt.Errorf("%s", "контейнер "+image+" не запущен. Поднимите Selenium Grid и стартуйте контейнер test-node перед выполнением тестов")
	}

	args := []string{
		"exec",
		// запуск от имени seluser (если нужно, можно убрать или заменить)
		"-u", "seluser",
		"-w", "/app",
		"-e", "PYTHONPATH=/app",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		image,
		"pytest", "-q", "-v", "-s", "--junitxml=/app/results/results.xml",
	}

	fmt.Printf(">>> Запуск команды: docker %v\n", args)
	// if err := runCmd("docker", args...); err != nil {
	// 	fmt.Errorf("ошибка выполнения pytest в контейнере "+image+" : %w", err)
	// }

	//fmt.Printf("%s", "Результаты тестов сохранены в: /"+flowName+"/results.xml \n" /*filepath.Join(absResultsPath, "results.xml")*/)

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
	fmt.Printf(">>> docker cp %s %s\n", filepath.Join(dir, hostFile), dst)
	if err := runCmd("docker", "cp", filepath.Join(dir, hostFile), dst); err != nil {
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
	return runCmd("docker", "cp",
		container+":/app/results/results.xml",
		filepath.Join(hostPath, "results.xml"),
	)
}