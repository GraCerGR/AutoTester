package createconteinerpackage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

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

func DockerBuild(imageTag string, dockerfileName string, contextPath string) error {
	fmt.Printf(">>> Проверка и сборка образа: %s\n", imageTag)

	cmdCheck := exec.Command("docker", "images", "-q", imageTag)
	output, err := cmdCheck.Output()
	if err != nil {
		return fmt.Errorf("Не удалось проверить наличие образа: %w", err)
	}

	if len(output) > 0 {
		fmt.Println("Образ уже существует, сборка не требуется.")
		return nil
	}

	// Если образа нет, строим его
	fmt.Println("Образ не найден. Строим...")

	buildDir, err := filepath.Abs(contextPath)
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь: %w", err)
	}

	err = runCmd("docker", "build",
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
	return runCmd("docker", "run", "-d",
		"--name", containerName,
		"--network", "tests-net",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		imageTag+":latest",
		"sleep", "infinity",
	)
}

func RemoveTestContainer(containerName string) error {
	return runCmd("docker", "rm", "-f", containerName)
}
