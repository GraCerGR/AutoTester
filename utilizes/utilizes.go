package utilizes

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

func GenerateStringId() string {

	//now := time.Now()
	//timeStr := now.Format("2006-01-02_15-04-05")
	
	id := fmt.Sprintf("%s", uuid.New())
	return id
}


func CountingFolder(folderPath string) ([]int, map[int]string, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		msg := fmt.Sprintf("Не удалось прочитать каталог %s: %v\n", folderPath, err)
		return nil, nil, fmt.Errorf("%s", msg)
	}

	var dirs []int
	dirMap := make(map[int]string)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()

		match := regexp.MustCompile(`\d+`).FindString(name)
		if match == "" {
			continue
		}
		n, _ := strconv.Atoi(match)
		dirs = append(dirs, n)
		dirMap[n] = filepath.Join(folderPath, name)
	}

	if len(dirs) == 0 {
		msg := "Не найдено подпапок с числовыми именами."
		return nil, nil, fmt.Errorf("%s", msg)
	}
	return dirs, dirMap, nil
}
