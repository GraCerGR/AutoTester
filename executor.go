package main

import (
	"MainApp/checker"
	"MainApp/classes"
	"MainApp/commandongit"
	"MainApp/commandonhost"
	"MainApp/createconteinerpackage"
	myredis "MainApp/redis"

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

func Executor(ctx context.Context, redisClient *redis.Client, attempt classes.Attempt, containerTestName, containerSiteName string) {
	// ---- Executor ----

	// Поиск свободного тестового контейнера
	// containerTestName, err := myredis.GetFreeContainer(ctx, redisClient, settings.TestContainers, attempt.ProgrammingLanguageName)
	// if err != nil {
	// 	fmt.Println("Нет свободных контейнеров для запуска проекта, нужно подождать")
	// 	//ЖДАТЬ КОГДА ОСВОБОДИТСЯ, НЕ РЕТУРН
	// 	return
	// }
	// // Поиск свободного сайтового контейнера
	// containerSiteName, err := myredis.GetFreeContainer(ctx, redisClient, settings.SiteContainers, "")
	// if err != nil {
	// 	fmt.Println("Нет свободных контейнеров для запуска сайта, нужно подождать")
	// 	//ЖДАТЬ КОГДА ОСВОБОДИТСЯ, НЕ РЕТУРН
	// 	return
	// }

	fmt.Printf("Контейнеры выбраны для проверки:%v и %v\n", containerTestName, containerSiteName)

	//flolwName := utilizes.GenerateUniqueFolder()

	//Загрузка решения из гита
	if err := commandongit.DownloadFromGit(attempt.GitStudentURL, "main", "", "Solutions/Gits/"+containerTestName, ""); err != nil {
		fmt.Printf("Ошибка загрузки решения с гита: %v\n", err)
		return
	}

	//Передаём файлы теста в тестовый контейнер
	if err := createconteinerpackage.RunTestContainer(containerTestName); err != nil {
		fmt.Println("Не удалось создать test container:", err)
		return
	}

	//Запускается при получении из очереди запроса на проверку - получает решение студента
	if err := createconteinerpackage.SendProjectToImage("Solutions/Gits/"+containerTestName, containerTestName, false); err != nil {
		fmt.Printf("Ошибка передачи проекта в контейнер: %v\n", err)
		return
	}

	if err := createconteinerpackage.ReplaceTestURLInPythonContainer(containerTestName, attempt.VariableWithURL, "http://"+containerSiteName+":80"); err != nil { //"TEST_URL"
		fmt.Printf("Ошибка замены переменной "+attempt.VariableWithURL+": %v\n", err)
	}

	//Отправляет скрипт перенаправления запросов в хаб селениум грида
	if err := createconteinerpackage.SendSiteCustomize("DockerFiles/sitecustomize.py", containerTestName); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	//Загрузка сайта из гита
	if err := commandongit.DownloadFromGit(attempt.GitSiteURL, "main", "", "Sites/Gits/"+containerTestName, ""); err != nil {
		fmt.Printf("Ошибка загрузки сайта с гита: %v\n", err)
		return
	}

	// Папка потока - генерируется с именем тестового контейнера
	if _, err := ExecutionSolutionOnSites("Sites/Gits/"+containerTestName+"/Sites", "Sites/Gits/"+containerTestName+"/StudentResults", "Sites/Gits/"+containerTestName+"/Results", containerTestName, containerSiteName); err != nil {
		fmt.Printf("Ошибка при запуске автотестов в контейнере: %v\n", err)
		return
	}

	ClearAllContainers(containerTestName, containerSiteName)

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

		result, err := checker.CheckerWithNames(index, resultsFolder, correctResultsFolder)
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

	//checkerResult.TestingVerdict = classes.TestVerdictEnum.Ok
	b, _ := json.MarshalIndent(checkerResult, "", "  ")
	fmt.Println(string(b))

	return checkerResult, nil
}

// ---- Отчистка ----
func ClearAllContainers(containerTestName, containerSiteName string) {

	//Удаляем файлы теста в контейнере
	if err := createconteinerpackage.RemoveProjectFromContainer(containerTestName, false); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта в контейнере
	if err := createconteinerpackage.RemoveProjectFromContainer(containerSiteName, true); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта и решения с хоста
	if err := commandonhost.ClearHostFolder("Sites/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка отчистки файлов гита: %v\n", err)
	}
	if err := commandonhost.ClearHostFolder("Solutions/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка отчистки файлов гита: %v\n", err)
	}

	createconteinerpackage.RemoveTestContainer(containerTestName)
}
