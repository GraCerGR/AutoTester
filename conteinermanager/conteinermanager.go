package conteinermanager

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func RunCmd(ctx context.Context, name string, args ...string) error {
	fmt.Printf("[CMD] Выполнение: %s %v\n", name, args)

	cmd := exec.CommandContext(ctx, name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Ошибка запуска команды: %w", err)
	}

	go stream(stdout, "[OUT]")
	go stream(stderr, "[ERR]")

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("Ошибка выполнения команды: %w", err)
	}

	fmt.Println("[OK]")
	return nil
}

func RunCmdOutput(ctx context.Context, name string, args ...string) (string, error) {
	fmt.Printf("[CMD] Выполнение: %s %v\n", name, args)

	cmd := exec.CommandContext(ctx, name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Ошибка запуска команды: %w", err)
	}

	var buf bytes.Buffer
	go streamCopy(stdout, &buf)
	go streamCopy(stderr, &buf)

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	if err := cmd.Wait(); err != nil {
		return buf.String(), fmt.Errorf("Ошибка выполнения команды: %w", err)
	}

	return buf.String(), nil
}

func streamCopy(r io.Reader, buf *bytes.Buffer) {
	io.Copy(io.MultiWriter(buf, os.Stdout), r)
}

func runCmdForAutotests(ctx context.Context, logFilePath string, name string, args ...string) (bool, error) {
	fmt.Printf("[CMD] Выполнение: %s %v\n", name, args)

	cmd := exec.CommandContext(ctx, name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return false, fmt.Errorf("ошибка создания лог-файла %s: %w", logFilePath, err)
	}
	defer logFile.Close()

	stdoutWriter := io.MultiWriter(os.Stdout, logFile)
	stderrWriter := io.MultiWriter(os.Stderr, logFile)

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("Ошибка запуска команды: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(stdoutWriter, stdout)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(stderrWriter, stderr)
	}()

	err = cmd.Wait()
	wg.Wait()

	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err == nil {
		fmt.Println("[OK] (ошибка разрешена)")
		return true, nil
	}

	if _, ok := err.(*exec.ExitError); ok {
		fmt.Println("[OK] (ошибка разрешена):", err)
		return false, nil
	}

	return false, fmt.Errorf("Ошибка выполнения команды: %w", err)
}

func stream(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fmt.Printf("%s %s\n", prefix, scanner.Text())
	}
}

func checkContainerRunning(containerName string) (bool, error) {
	checkCmd := exec.Command(
		"docker", "ps", "-q",
		"-f", fmt.Sprintf("name=%s", containerName),
	)

	out, err := checkCmd.Output()
	if err != nil {
		return false, fmt.Errorf("не удалось выполнить проверку контейнеров: %w", err)
	}

	if len(bytes.TrimSpace(out)) == 0 {
		return false, fmt.Errorf(
			"Контейнер %s не запущен",
			containerName,
		)
	}
	return true, nil
}

func DockerBuild(ctx context.Context, imageTag string, dockerfileName string, contextPath string) error {
	fmt.Printf("Проверка и сборка образа: %s\n", imageTag)

	cmdCheck := exec.Command("docker", "images", "-q", imageTag)
	output, err := cmdCheck.Output()
	if err != nil {
		return fmt.Errorf("Не удалось проверить наличие образа: %w", err)
	}

	if len(output) > 0 {
		fmt.Println("Образ уже существует, сборка не требуется.")
		return nil
	}

	fmt.Println("Образ не найден. Строим...")

	buildDir, err := filepath.Abs(contextPath)
	if err != nil {
		return fmt.Errorf("Не удалось получить абсолютный путь: %w", err)
	}

	err = RunCmd(ctx, "docker", "build",
		"-t", imageTag+":latest",
		"-f", dockerfileName,
		buildDir,
	)
	if err != nil {
		return fmt.Errorf("Не удалось построить образ: %w", err)
	}

	fmt.Println("Образ успешно построен:", imageTag)
	return nil
}

func RunTestContainer(ctx context.Context, containerName string, imageTag string) error {
	return RunCmd(ctx, "docker", "run", "-d",
		"--name", containerName,
		"--network", "tests-net",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		"-e", "SESSION_NAME="+containerName,
		imageTag+":latest",
		"sleep", "infinity",
	)
}

func RunSiteContainer(ctx context.Context, containerName string, imageTag string) error {
	return RunCmd(ctx, "docker", "run", "-d",
		"--name", containerName,
		"--network", "tests-net",
		"--restart", "unless-stopped",
		imageTag+":latest",
	)
}

func RemoveContainer(ctx context.Context, containerName string) error {
	return RunCmd(ctx, "docker", "rm", "-f", containerName)
}

func SendProjectToImage(ctx context.Context, contextDir, containerName string, site bool) error {
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

	if err := RunCmd(ctx, "docker", "cp", hostSrc, dest); err != nil {
		return fmt.Errorf("ошибка копирования проекта в контейнер %s: %w", containerName, err)
	}

	if site {
		if err := RunCmd(ctx, "docker", "exec", "-u", "root", containerName, "chown", "-R", "nginx:nginx", "/usr/share/nginx/html"); err != nil {
			return fmt.Errorf("ошибка установки прав в контейнере: %w", err)
		}
	} else {
		if err := RunCmd(ctx, "docker", "exec", "-u", "root", containerName, "chown", "-R", "seluser:seluser", "/app"); err != nil {
			return fmt.Errorf("ошибка установки прав в контейнере: %w", err)
		}

	}
	if site {
		fmt.Println("Сайт успешно отправлен в контейнер", containerName)
	} else {
		fmt.Println("Решение успешно отправлено в контейнер", containerName)
	}

	return nil
}

func RemoveProjectFromContainer(ctx context.Context, containerName string, site bool) error {
	if containerName == "" {
		return fmt.Errorf("containerName must be provided")
	}

	var targetDir, preserveFile string
	if site {
		targetDir = "/usr/share/nginx/html"
		preserveFile = "50x.html"
		fmt.Printf("Удаляем файлы из %s в контейнере %s, кроме %s\n", targetDir, containerName, preserveFile)
	} else {
		targetDir = "/app"
		fmt.Printf("Удаляем файлы из %s в контейнере %s\n", targetDir, containerName)
	}

	if isRunning, err := checkContainerRunning(containerName); err != nil || !isRunning {
		return fmt.Errorf("Контейнер %s не запущен: %w", containerName, err)
	}

	// Удаляем содержимое каталога от имени пользователя seluser
	var rmCmd string
	if site {
		rmCmd = fmt.Sprintf("find %s -mindepth 1 ! -name '%s' -exec rm -rf {} +", targetDir, preserveFile)
	} else {
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

	if err := RunCmd(ctx, "docker", args...); err != nil {
		return fmt.Errorf("Ошибка удаления проекта в контейнере %s: %w", containerName, err)
	}

	fmt.Println("Файлы проекта успешно удалены из контейнера", containerName)
	return nil
}

func AddSiteRedirectToContainer(ctx context.Context, testContainer, testURL, targetContainer, network string) error {

	u, err := url.Parse(testURL)
	if err != nil {
		return fmt.Errorf("не удалось распарсить %s: %w", testURL, err)
	}
	host := u.Hostname()

	out, err := RunCmdOutput(ctx, "docker", "inspect", "-f", fmt.Sprintf("{{index .NetworkSettings.Networks \"%s\" \"IPAddress\"}}", network), targetContainer)
	if err != nil {
		return fmt.Errorf("не удалось получить IP адрес контейнера %s: %w", targetContainer, err)
	}

	ip := strings.TrimSpace(out)
	if ip == "" {
		return fmt.Errorf("IP контейнера %s пустой", targetContainer)
	}

	args := []string{
		"exec",
		"-u", "root",
		testContainer,
		"bash", "-c",
		fmt.Sprintf("echo '%s %s' >> /etc/hosts", ip, host),
	}

	if err := RunCmd(ctx, "docker", args...); err != nil {
		return fmt.Errorf("Не удалось добавить перенаправление: %w", err)
	}

	fmt.Printf("Перенаправление %s (%s) -> %s успешно добавлено в контейнер\n", testURL, ip, targetContainer)

	return nil
}
