package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFromGit(ctx context.Context, repoURL, branch, subDir, targetDir, token string) error {
	if repoURL == "" {
		return fmt.Errorf("RepoURL is required")
	}
	if targetDir == "" {
		return fmt.Errorf("TargetDir is required")
	}
	if branch == "" {
		branch = "main"
	}

	zipURL, err := buildZipURL(repoURL, branch)
	if err != nil {
		return err
	}

	fmt.Println("Загрузка репозитория:", zipURL)

	req, err := http.NewRequestWithContext(ctx, "GET", zipURL, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Ошибка загрузки репозитория: %s", resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return err
	}

	os.RemoveAll(targetDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	var root string

	for _, f := range reader.File {
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if root == "" {
			root = strings.Split(f.Name, "/")[0]
		}

		rel := strings.TrimPrefix(f.Name, root+"/")

		if subDir != "" {
			if !strings.HasPrefix(rel, subDir+"/") {
				continue
			}
			rel = strings.TrimPrefix(rel, subDir+"/")
		}

		if rel == "" {
			continue
		}

		dst := filepath.Join(targetDir, rel)

		if f.FileInfo().IsDir() {
			os.MkdirAll(dst, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(dst)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()

		if err != nil {
			return err
		}
	}

	fmt.Println("Загрузка репозитория завершена в:", targetDir)
	return nil
}

func buildZipURL(repo, ref string) (string, error) {
	repo = strings.TrimSuffix(repo, ".git")

	if strings.Contains(repo, "github.com") {
		return fmt.Sprintf("%s/archive/refs/heads/%s.zip", repo, ref), nil
	}

	if strings.Contains(repo, "gitlab.com") {
		return fmt.Sprintf("%s/-/archive/%s/repo-%s.zip", repo, ref, ref), nil
	}

	if strings.Contains(repo, "bitbucket.org") {
		return fmt.Sprintf("%s/get/%s.zip", repo, ref), nil
	}

	return "", fmt.Errorf("unsupported git provider")
}
