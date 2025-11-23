package main

import (
	"MainApp/classes"
	"MainApp/createconteinerpackage"
	selenium_grid "MainApp/seleniumgrid"
	"MainApp/utilizes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"MainApp/checker"
	"fmt"
)

func main() {

	//Запуск компоуза с гридом, тестом и сайтом
	//Запускается при запуске всего хаба
	if err := selenium_grid.StartSeleniumGrid("./seleniumgrid"); err != nil {
		fmt.Printf("Ошибка запуска Selenium Grid: %v\n", err)
		return
	}

	//Передаём файлы теста в тестовый контейнер
	//Запускается при получении из очереди запроса на проверку - получает решение студента
	if err := createconteinerpackage.SendProjectToImage("SolutionTests/ExampleSite", "test-node", false); err != nil {
		fmt.Printf("Ошибка передачи проекта студента в контейнер: %v\n", err)
		return
	}

	//Отправляет скрипт перенаправления запросов в хаб селениум грида
	if err := createconteinerpackage.SendSiteCustomize("DockerFiles/sitecustomize.py", "test-node"); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	//ExampleSite - запрашиваю из бд или гита
	// Папка потока - генерируется с уникальным id, чтобы потоки отличались
	ExecutionSolutionOnSites("Sites/ExampleSite", "results/studentResults/"+utilizes.GenerateUniqueFolder(), "results/correctResults/ExampleSite")

	//Удаляем файлы теста в контейнере
	if err := createconteinerpackage.RemoveProjectFromContainer("test-node", false); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
		return
	}

	fmt.Printf("OK")
}

func ExecutionSolutionOnSites(siteFolder, resultsFolder, standartFolder string) (checker.AllTestsInChecker, error) {
	var checkerResult checker.AllTestsInChecker

	entries, err := os.ReadDir(siteFolder)
	if err != nil {
		msg := fmt.Sprintf("Не удалось прочитать каталог %s: %v\n", siteFolder, err)
		checkerResult.Comment = msg
		return checkerResult, err
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
		dirMap[n] = filepath.Join(siteFolder, name)
	}

	if len(dirs) == 0 {
		msg := fmt.Sprintf("Не найдено подпапок с числовыми именами.")
		checkerResult.Comment = msg
		return checkerResult, err
	}

	for _, index := range dirs {
		path := dirMap[index]

		if err := createconteinerpackage.SendProjectToImage(path, "site", true); err != nil {
			msg := fmt.Sprintf("Ошибка при отправке %s: %v\n", path, err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		//Запускает тесты в контейнере
		if err := createconteinerpackage.RunPythonTestsContainer("test-node", resultsFolder); err != nil {
			msg := fmt.Sprintf("Ошибка выполнения тестов: %v\n", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		//Вывод реузльтатов теста в json
		if err := checker.Parsing(resultsFolder+"/results.xml", index); err != nil {
			msg := fmt.Sprintf("Ошибка вывода результатов в json: %v\n", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		result, err := checker.Checker(index, resultsFolder, standartFolder)
		if err != nil {
			msg := fmt.Sprintf("Ошибка сравнения:", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		checkerResult.AllTests = append(checkerResult.AllTests, result)

		//Удаляем файлы сайта в сайтовом контейнере
		if err := createconteinerpackage.RemoveProjectFromContainer("site", true); err != nil {
			msg := fmt.Sprintf("Ошибка при очистке контейнера: %v\n", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		lastVerdict := checkerResult.AllTests[len(checkerResult.AllTests)-1].TestingVerdict
		lastComment := checkerResult.AllTests[len(checkerResult.AllTests)-1].Comment
		if lastVerdict != classes.TestVerdictEnum.Ok {
			checkerResult.TestingVerdict = lastVerdict
			checkerResult.Comment = lastComment
			break
		}
	}

	checkerResult.TestingVerdict = classes.TestVerdictEnum.Ok
	b, _ := json.MarshalIndent(checkerResult, "", "  ")
	fmt.Println(string(b))

	return checkerResult, nil
}
