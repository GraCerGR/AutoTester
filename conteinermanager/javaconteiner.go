package conteinermanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunJavaTestsContainer(containerName string) (bool, error) {

	if isRunning, err := checkContainerRunning(containerName); err != nil || !isRunning {
		return false, fmt.Errorf("контейнер %s не запущен: %w", containerName, err)
	}

	args := []string{
		"exec",
		"-u", "seluser",
		"-w", "/app",
		"-e", "SELENIUM_HUB=http://selenium-hub:4444",
		containerName,
		"mvn", "clean", "test",
	}

	fmt.Printf("Запуск Java тестов: docker %v\n", args)

	passed, err := runCmdAllowFail("docker", args...)
	if err != nil {
		return false, err
	}
	if passed {
		fmt.Println("Java тесты прошли")
	} else {
		fmt.Println("Java тесты не прошли (но продолжаем)")
	}

	return passed, nil
}

func SendSiteJavaCustomize(dir, hostFile, containerName string) error {
	if hostFile == "" {
		return fmt.Errorf("hostFile must be provided")
	}

	hostPath := filepath.Join(dir, hostFile)
	if _, err := os.Stat(hostPath); err != nil {
		return fmt.Errorf("файл не найден: %w", err)
	}

	// Целевая директория внутри контейнера
	containerDir := "/app/src/test/java/org/selenium/chrome/"
	dst := fmt.Sprintf("%s:%s%s", containerName, containerDir, hostFile)

	if err := RunCmd("docker", "exec", containerName, "mkdir", "-p", containerDir); err != nil {
		return fmt.Errorf("не удалось создать директорию в контейнере: %w", err)
	}

	if err := RunCmd("docker", "cp", hostPath, dst); err != nil {
		return fmt.Errorf("ошибка копирования: %w", err)
	}

	fmt.Println(hostFile + " скопирован в " + containerDir + " контейнера")
	return nil
}

func ReplaceTestURLInJavaContainer(containerName, varName, newURL string) error {
	// Экранируем символы для sed
	escapedURL := strings.NewReplacer(
		`&`, `\&`,
		`|`, `\|`,
		`\`, `\\`,
	).Replace(newURL)

	// sed: заменяем строки вида:
	// ... TEST_URL = "..."
	sedExpr := fmt.Sprintf(`s|\(%s\s*=\s*\)".*"|\1"%s"|`, varName, escapedURL)

	cmd := exec.Command(
		"docker", "exec", containerName,
		"bash", "-c",
		fmt.Sprintf(`find /app -name '*.java' -exec sed -i '%s' {} +`, sedExpr),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func AddSurefireReportConfig(projectDir, resultsDir string) error {
	if resultsDir == "" {
		resultsDir = "results"
	}

	// Находим pom.xml
	var pomPath string
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "pom.xml" {
			pomPath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ошибка поиска pom.xml: %w", err)
	}
	if pomPath == "" {
		return fmt.Errorf("pom.xml не найден в %s", projectDir)
	}

	// Читаем файл
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать pom.xml: %w", err)
	}

	strContent := string(content)

	surefireConfig := fmt.Sprintf(`
        <configuration>
            <reportsDirectory>${project.basedir}/%s</reportsDirectory>
            <useFile>true</useFile>
        </configuration>
`, resultsDir)

	// 1Если плагин уже есть
	if strings.Contains(strContent, "<artifactId>maven-surefire-plugin</artifactId>") {
		// Вставим конфигурацию перед закрывающим </plugin>
		strContent = strings.Replace(strContent, "</plugin>", surefireConfig+"</plugin>", 1)
	} else {
		// 2Если <plugins> есть, но плагина нет
		if strings.Contains(strContent, "<plugins>") {
			pluginBlock := fmt.Sprintf(`
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-surefire-plugin</artifactId>
            <version>3.2.5</version>%s
        </plugin>
`, surefireConfig)
			strContent = strings.Replace(strContent, "<plugins>", "<plugins>"+pluginBlock, 1)
		} else {
			// Если <plugins> нет, создаём его внутри <build>
			pluginBlock := fmt.Sprintf(`
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-surefire-plugin</artifactId>
            <version>3.2.5</version>%s
        </plugin>
    </plugins>
</build>
`, surefireConfig)
			strContent = strings.Replace(strContent, "</build>", pluginBlock, 1)
		}
	}

	// Записываем обратно
	err = os.WriteFile(pomPath, []byte(strContent), 0644)
	if err != nil {
		return fmt.Errorf("не удалось записать pom.xml: %w", err)
	}

	fmt.Println("pom.xml успешно обновлён для записи XML-отчётов в", resultsDir)
	return nil
}

func RenameSurefireXML(resultsDir string) error {
	// ищем файлы TEST-*.xml
	matches, err := filepath.Glob(filepath.Join(resultsDir, "TEST-*.xml"))
	if err != nil {
		return fmt.Errorf("ошибка поиска XML файлов: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("файл TEST-*.xml не найден в %s", resultsDir)
	}

	src := matches[0]
	dst := filepath.Join(resultsDir, "results.xml")

	// переименовываем
	err = os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("не удалось переименовать %s в %s: %w", src, dst, err)
	}

	fmt.Println("XML файл переименован в", dst)
	return nil
}

func CopyResultsFromJavaContainer(container, hostPath string) error {
	os.MkdirAll(hostPath, 0755)
	cmd := exec.Command("docker", "exec", container, "sh", "-c", "ls /app/target/surefire-reports/TEST-*.xml | head -n 1")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("не удалось найти TEST-*.xml в контейнере: %w", err)
	}

	srcFile := strings.TrimSpace(string(out))
	if srcFile == "" {
		return fmt.Errorf("TEST-*.xml файл не найден в контейнере")
	}

	dstFile := filepath.Join(hostPath, "results.xml")

	if err := RunCmd("docker", "cp", fmt.Sprintf("%s:%s", container, srcFile), dstFile); err != nil {
		return fmt.Errorf("не удалось скопировать файл из контейнера: %w", err)
	}

	fmt.Println("Файл результатов скопирован как", dstFile)
	return nil
}