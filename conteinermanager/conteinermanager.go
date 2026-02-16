package conteinermanager

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func RunCmd(name string, args ...string) error {
	fmt.Printf("[CMD] Выполнение: %s %v\n", name, args)

	cmd := exec.Command(name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Ошибка запуска команды: %w", err)
	}

	go stream(stdout, "[OUT]")
	go stream(stderr, "[ERR]")

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("Ошибка выполнения команды: %w", err)
	}

	fmt.Println("[OK] УСПЕХ")
	return nil
}

func runCmdAllowFail(name string, args ...string) (bool, error) {
	fmt.Printf("[CMD] Выполнение: %s %v\n", name, args)

	cmd := exec.Command(name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("Ошибка запуска команды: %w", err)
	}

	go stream(stdout, "[OUT]")
	go stream(stderr, "[ERR]")

	err := cmd.Wait()

	if err == nil {
		fmt.Println("[OK] УСПЕХ (ошибка разрешена)")
		return true, nil
	}

	if _, ok := err.(*exec.ExitError); ok {
		fmt.Println("[OK] УСПЕХ (ошибка разрешена):", err)
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
			"Контейнер %s не запущен. Поднимите Selenium Hub и стартуйте контейнер перед выполнением тестов",
			containerName,
		)
	}
	return true, nil
}

func DockerBuild(imageTag string, dockerfileName string, contextPath string) error {
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

	err = RunCmd("docker", "build",
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

func dockerRun(imageTag, name string, hostPort int) error {
	_ = exec.Command("docker", "rm", "-f", name).Run()

	portMapping := fmt.Sprintf("%d:80", hostPort)
	cmd := exec.Command("docker", "run", "-d", "--name", name, "-p", portMapping, imageTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}

func RunTestContainer(containerName string, imageTag string) error {
	return RunCmd("docker", "run", "-d",
		"--name", containerName,
		"--network", "tests-net",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		imageTag+":latest",
		"sleep", "infinity",
	)
}

func RemoveTestContainer(containerName string) error {
	return RunCmd("docker", "rm", "-f", containerName)
}

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

	if err := RunCmd("docker", "cp", hostSrc, dest); err != nil {
		return fmt.Errorf("ошибка копирования проекта в контейнер %s: %w", containerName, err)
	}

	if site {
		if err := RunCmd("docker", "exec", "-u", "root", containerName, "chown", "-R", "nginx:nginx", "/usr/share/nginx/html"); err != nil {
			return fmt.Errorf("ошибка установки прав в контейнере: %w", err)
		}
	} else {
		if err := RunCmd("docker", "exec", "-u", "root", containerName, "chown", "-R", "seluser:seluser", "/app"); err != nil {
			return fmt.Errorf("ошибка установки прав в контейнере: %w", err)
		}

	}

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

	if err := RunCmd("docker", args...); err != nil {
		return fmt.Errorf("Ошибка удаления проекта в контейнере %s: %w", containerName, err)
	}

	fmt.Println("Файлы проекта успешно удалены из контейнера", containerName)
	return nil
}
