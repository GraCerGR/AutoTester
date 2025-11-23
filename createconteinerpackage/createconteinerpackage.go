package createconteinerpackage

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func CreateConteinerWithSite(sitePath, imageTag, containerName string, hostPort int) {

	if err := ensureSiteExists(sitePath); err != nil {
		log.Fatalf("Файлы сайта не найдены: %v", err)
	}

	if err := checkDockerAvailable(); err != nil {
		log.Fatalf("Docker недоступен: %v", err)
	}

	ctxDir, err := os.MkdirTemp("", "site-build-")
	if err != nil {
		log.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(ctxDir)

	if err := createBuildContext(sitePath, ctxDir); err != nil {
		log.Fatalf("Не удалось подготовить build context: %v", err)
	}

	fmt.Println("Сборка образа с сайтом...")
	if err := dockerBuild(ctxDir, imageTag); err != nil {
		log.Fatalf("Ошибка при docker build: %v", err)
	}

	fmt.Println("Запуск контейнера...")
	if err := dockerRun(imageTag, containerName, hostPort); err != nil {
		log.Fatalf("Ошибка при docker run: %v", err)
	}

	fmt.Printf("Готово! Сайт доступен по адресу: http://localhost:%d\n", hostPort)
}

func ensureSiteExists(path string) error {
	_, err := os.Stat(path)
	return err
}

func checkDockerAvailable() error {
	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("команда 'docker' недоступна или демон не запущен: %w", err)
	}
	return nil
}

func createBuildContext(sitePath, ctxDir string) error {
	info, err := os.Stat(sitePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(sitePath)
		if err != nil {
			return err
		}
		for _, e := range entries {
			srcPath := filepath.Join(sitePath, e.Name())
			dstPath := filepath.Join(ctxDir, e.Name())

			if e.IsDir() {
				if err := copyDir(srcPath, dstPath); err != nil {
					return err
				}
			} else {
				in, err := os.Open(srcPath)
				if err != nil {
					return err
				}
				defer in.Close()

				out, err := os.Create(dstPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(out, in); err != nil {
					out.Close()
					return err
				}
				out.Close()
			}
		}
	} else {
		// если передан один файл, положить как index.html
		in, err := os.Open(sitePath)
		if err != nil {
			return err
		}
		defer in.Close()
		outPath := filepath.Join(ctxDir, "index.html")
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
	}

	// Dockerfile
	dockerfile := `FROM nginx:alpine
					COPY . /usr/share/nginx/html
					EXPOSE 80
					CMD ["nginx", "-g", "daemon off;"]`
	dfPath := filepath.Join(ctxDir, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(dockerfile), 0644); err != nil {
		return err
	}

	return nil
}

// copyDir рекурсивно копирует содержимое src в dst.
// Создаёт директории и файлы, копирует содержимое и права.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			// создаём директорию
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			return nil
		}

		// файл — копируем
		info, err := d.Info()
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// убедимся, что родительская директория существует
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		out.Close()

		// применяем права (не все файловые системы поддерживают точные права, но попытаться стоит)
		if err := os.Chmod(target, info.Mode()); err != nil {
			log.Printf("Не удалось установить права для %s: %v", target, err)
		}
		return nil
	})
}

func dockerBuild(ctxDir, imageTag string) error {
	cmd := exec.Command("docker", "build", "-t", imageTag, ".")
	cmd.Dir = ctxDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
