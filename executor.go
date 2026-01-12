package main

import (
	"MainApp/checker"
	"MainApp/classes"
	"MainApp/commandonhost"
	"MainApp/createconteinerpackage"
	myredis "MainApp/redis"
	"MainApp/settings"

	//"MainApp/utilizes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func Executor(ctx context.Context, redisClient *redis.Client, attempt classes.Attempt) {
	// ---- Executor ----

	// Поиск свободного тестового контейнера
	containerTestName, err := myredis.GetFreeContainer(ctx, redisClient, settings.TestContainers, attempt.ProgrammingLanguageName)
	if err != nil {
		fmt.Println("Нет свободных контейнеров для запуска проекта, нужно подождать")
		//ЖДАТЬ КОГДА ОСВОБОДИТСЯ, НЕ РЕТУРН
		return
	}
	// Поиск свободного сайтового контейнера
	containerSiteName, err := myredis.GetFreeContainer(ctx, redisClient, settings.SiteContainers, "")
	if err != nil {
		fmt.Println("Нет свободных контейнеров для запуска сайта, нужно подождать")
		//ЖДАТЬ КОГДА ОСВОБОДИТСЯ, НЕ РЕТУРН
		return
	}

	myredis.SetContainerStatus(ctx, redisClient, containerTestName, "busy")
	myredis.SetContainerStatus(ctx, redisClient, containerSiteName, "busy")

	fmt.Printf("Контейнеры выбраны для проверки:%v и %v\n", containerTestName, containerSiteName)

	//flolwName := utilizes.GenerateUniqueFolder()

	//Передаём файлы теста в тестовый контейнер
	//Запускается при получении из очереди запроса на проверку - получает решение студента
	if err := createconteinerpackage.SendProjectToImage("SolutionTests/ExampleSite", containerTestName, false); err != nil {
		fmt.Printf("Ошибка передачи проекта студента в контейнер: %v\n", err)
		return
	}

	if err := createconteinerpackage.ReplaceTestURLInPythonContainer(containerTestName, attempt.VariableWithURL, "http://"+containerSiteName+":80"); err != nil { //"TEST_URL"
		fmt.Printf("Ошибка замены TEST_URL: %v\n", err)
	}

	//Отправляет скрипт перенаправления запросов в хаб селениум грида
	if err := createconteinerpackage.SendSiteCustomize("DockerFiles/sitecustomize.py", containerTestName); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	//ExampleSite - запрашиваю из гита
	if err := createconteinerpackage.DownloadFromGit(attempt.GitSiteURL, "main", "", "Sites/Gits/"+containerTestName, ""); err != nil {
		fmt.Printf("Ошибка загрузки сайта: %v\n", err)
		return
	}
	// Папка потока - генерируется с уникальным id, чтобы потоки отличались - теперь имя тестового контейнера
	if _, err := ExecutionSolutionOnSites("Sites/Gits/"+containerTestName+"/Sites", "Sites/Gits/"+containerTestName+"/StudentResults", "Sites/Gits/"+containerTestName+"/Results", containerTestName, containerSiteName); err != nil {
		fmt.Printf("Ошибка при запуске автотестов в контейнере: %v\n", err)
		return
	}

	//Удаляем файлы теста в контейнере
	if err := createconteinerpackage.RemoveProjectFromContainer(containerTestName, false); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
		return
	}

	//Удаляем файлы сайта с хоста
	if err := commandonhost.ClearHostFolder("Sites/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка отчистки файлов гита: %v\n", err)
		return
	}

	myredis.SetContainerStatus(ctx, redisClient, containerTestName, "free")
	myredis.SetContainerStatus(ctx, redisClient, containerSiteName, "free")

	fmt.Printf("OK")
}

func ExecutionSolutionOnSites(siteFolder, resultsFolder, correctResultsFolder, containerTest, conteinerSite string) (checker.AllTestsInChecker, error) {
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

		if err := createconteinerpackage.SendProjectToImage(path, conteinerSite, true); err != nil {
			msg := fmt.Sprintf("Ошибка при отправке %s: %v\n", path, err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		//Запускает тесты в контейнере
		if err := createconteinerpackage.RunPythonTestsContainer(containerTest, resultsFolder); err != nil {
			msg := fmt.Sprintf("Ошибка выполнения тестов: %v\n", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		//Загрузка результатов с контейнера
		if err := createconteinerpackage.CopyResultsFromContainer(containerTest, resultsFolder); err != nil {
			msg := fmt.Sprintf("Ошибка загрузки результатов с контейнера: %v\n", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		// ---- Checker ----

		//Вывод реузльтатов теста в json
		if err := checker.Parsing(resultsFolder+"/results.xml", index); err != nil {
			msg := fmt.Sprintf("Ошибка вывода результатов в json: %v\n", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		result, err := checker.Checker(index, resultsFolder, correctResultsFolder)
		if err != nil {
			msg := fmt.Sprintf("Ошибка сравнения:", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		checkerResult.AllTests = append(checkerResult.AllTests, result)

		//Удаляем файлы сайта в сайтовом контейнере
		if err := createconteinerpackage.RemoveProjectFromContainer(conteinerSite, true); err != nil {
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
