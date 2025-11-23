package createconteinerpackage

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func SendProjectToImage(contextDir, containerName string, site bool) error {
	// Валидация базовых параметров
	if contextDir == "" {
		return fmt.Errorf("contextDir must be provided")
	}

	// Приводим contextDir к абсолютному пути (чтобы docker cp корректно работал)
	absContextDir, err := filepath.Abs(contextDir)
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь для contextDir: %w", err)
	}

	// Проверка запущенного тестового контейнера
	checkCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	out, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("не удалось проверить статус контейнера %s: %w", containerName, err)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return fmt.Errorf("контейнер '%s' не запущен: невозможно скопировать файлы проекта. Поднимите Selenium Grid и контейнер перед выполнением тестов", containerName)
	}

	// Копируем содержимое contextDir в /app контейнера test-node
	hostSrc := filepath.Clean(absContextDir) + string(os.PathSeparator) + "."
	dest := fmt.Sprintf("%s:/app/", containerName)

	if site {
		dest = fmt.Sprintf("%s:/usr/share/nginx/html/", containerName)
	}

	fmt.Printf(">>> docker cp %s %s\n", hostSrc, dest)
	if err := runCmd("docker", "cp", hostSrc, dest); err != nil {
		return fmt.Errorf("ошибка копирования проекта в контейнер %s: %w", containerName, err)
	}

	if site {
		if err := runCmd("docker", "exec", "-u", "root", containerName, "chown", "-R", "nginx:nginx", "/usr/share/nginx/html"); err != nil {
			return fmt.Errorf("ошибка установки прав в контейнере: %w", err)
		}
	} else {
		if err := runCmd("docker", "exec", "-u", "root", containerName, "chown", "-R", "seluser:seluser", "/app"); err != nil {
			return fmt.Errorf("ошибка установки прав в контейнере: %w", err)
		}

	}

	return nil
}

func RunPythonTestsContainer(image, flowName string) error {
	// absResultsPath, err := filepath.Abs(projectPath)
	// if err != nil {
	// 	return fmt.Errorf("не удалось получить абсолютный путь для проекта: %w", err)
	// }

	// if err := os.MkdirAll(absResultsPath, 0o755); err != nil {
	// 	return fmt.Errorf("не удалось создать папку results: %w", err)
	// }

	// Проверим, что контейнер test-node запущен
	checkCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", image))
	out, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("не удалось выполнить проверку контейнеров: %w", err)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return fmt.Errorf("%s", "контейнер "+image+" не запущен. Поднимите Selenium Grid и стартуйте контейнер test-node перед выполнением тестов")
	}

	args := []string{
		"exec",
		// запуск от имени seluser (если нужно, можно убрать или заменить)
		"-u", "seluser",
		"-w", "/app",
		"-e", "PYTHONPATH=/app",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		image,
		"pytest", "-q", "-v", "-s", "--junitxml=/" + flowName + "/results.xml",
	}

	fmt.Printf(">>> Запуск команды: docker %v\n", args)
	if err := runCmd("docker", args...); err != nil {
		fmt.Errorf("ошибка выполнения pytest в контейнере "+image+" : %w", err)
	}

	fmt.Printf("%s", "Результаты тестов сохранены в: /"+flowName+"/results.xml \n" /*filepath.Join(absResultsPath, "results.xml")*/)

	return nil
}

// Не нужно
func RunSiteContainer(containerName string) error {
	// Проверяем, что контейнер запущен
	checkCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	out, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("не удалось проверить контейнер %s: %w", containerName, err)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return fmt.Errorf("контейнер '%s' не запущен: невозможно запустить сайт", containerName)
	}

	// Запускаем nginx в контейнере
	args := []string{
		"exec",
		"-u", "root",
		containerName,
		"nginx",
		"-g", "daemon off;",
	}

	fmt.Printf(">>> Запуск команды: docker %v\n", args)
	if err := runCmd("docker", args...); err != nil {
		return fmt.Errorf("ошибка запуска nginx в контейнере %s: %w", containerName, err)
	}

	fmt.Printf("Сайт запущен в контейнере %s\n", containerName)
	return nil
}

func RemoveProjectFromContainer(containerName string, site bool) error {
	if containerName == "" {
		return fmt.Errorf("containerName must be provided")
	}

	var targetDir, preserveFile string
	if site {
		targetDir = "/usr/share/nginx/html"
		preserveFile = "50x.html"
		fmt.Printf(">>> Удаляем файлы из %s в контейнере %s, кроме %s\n", targetDir, containerName, preserveFile)
	} else {
		targetDir = "/app"
		fmt.Printf(">>> Удаляем файлы из %s в контейнере %s\n", targetDir, containerName)
	}

	// Проверяем, что контейнер запущен
	checkCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	out, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("не удалось проверить статус контейнера %s: %w", containerName, err)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return fmt.Errorf("контейнер '%s' не запущен, нечего удалять", containerName)
	}

	// Удаляем содержимое каталога от имени пользователя seluser
	var rmCmd string
	if site {
		// Удаляем всё кроме 50x.html
		rmCmd = fmt.Sprintf("find %s -mindepth 1 ! -name '%s' -exec rm -rf {} +", targetDir, preserveFile)
	} else {
		// Удаляем всё в /app
		rmCmd = fmt.Sprintf("rm -rf %s/* %s/.[!.]* %s/..?*", targetDir, targetDir, targetDir)
	}

	args := []string{
		"exec",
		"-u", "root",
		containerName,
		"sh",
		"-c",
		rmCmd,
	}

	if err := runCmd("docker", args...); err != nil {
		return fmt.Errorf("ошибка удаления проекта в контейнере %s: %w", containerName, err)
	}

	fmt.Println("Файлы проекта успешно удалены из контейнера", containerName)
	return nil
}

func SendSiteCustomize(hostFile, containerName string) error {
	if hostFile == "" {
		return fmt.Errorf("hostFile must be provided")
	}
	if _, err := os.Stat(hostFile); err != nil {
		return fmt.Errorf("файл не найден: %w", err)
	}

	// Копируем в /app контейнера
	dst := fmt.Sprintf("%s:/app/sitecustomize.py", containerName)
	fmt.Printf(">>> docker cp %s %s\n", hostFile, dst)
	if err := runCmd("docker", "cp", hostFile, dst); err != nil {
		return fmt.Errorf("ошибка копирования: %w", err)
	}

	fmt.Println("sitecustomize.py скопирован в /app контейнера")
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
