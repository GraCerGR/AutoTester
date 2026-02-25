package main

import (
	"MainApp/checker"
	"MainApp/classes"
	"MainApp/commands"
	"MainApp/conteinermanager"
	myredis "MainApp/redis"
	"MainApp/settings"
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

	timeoutCtx, cancel := context.WithTimeout(ctx, settings.ExecutionTimeout)
	defer cancel()

	resultChan := make(chan classes.AllTestsInChecker)

	go func() {
        result := ExecutorMain(timeoutCtx, redisClient, attempt, containerTestName, containerSiteName)
        resultChan <- result
    }()

	var checkerResult classes.AllTestsInChecker

	select {
	case checkerResult = <-resultChan:
		b, _ := json.MarshalIndent(checkerResult, "", "  ")
		fmt.Println("Результат проверки:")
		fmt.Println(string(b))
	case <-timeoutCtx.Done():
		fmt.Println("Превышено время выполнения %d минут", settings.ExecutionTimeout)
		checkerResult.TestingVerdict = classes.TestVerdictEnum.Timeout
		checkerResult.Comment = fmt.Sprintf("Превышено время выполнения %d минут", settings.ExecutionTimeout)
	}

	// Отправка результата проверки

	ClearAllContainers(containerTestName, containerSiteName)

	myredis.SetContainerStatus(ctx, redisClient, containerTestName, "free")
	myredis.SetContainerStatus(ctx, redisClient, containerSiteName, "free")
	fmt.Printf("Executor свободен для новых задач\n")
}

func ExecutorMain(ctx context.Context, redisClient *redis.Client, attempt classes.Attempt, containerTestName, containerSiteName string) classes.AllTestsInChecker {
	result := classes.AllTestsInChecker{}

	fmt.Printf("Контейнеры выбраны для проверки:%v и %v\n", containerTestName, containerSiteName)

	//Загрузка решения из гита
	if err := commands.DownloadFromGit(attempt.GitStudentURL, "main", "", "Solutions/Gits/"+containerTestName, ""); err != nil {
		fmt.Printf("Ошибка загрузки решения с гита: %v\n", err)
		result.TestingVerdict = classes.TestVerdictEnum.Fail
		return result
	}

	//Запускаем нужный контейнер для тестов
	if err := conteinermanager.RunTestContainer(containerTestName, settings.ChooseImageTag(attempt.ProgrammingLanguageName)); err != nil {
		fmt.Println("Не удалось создать test container:", err)
		result.TestingVerdict = classes.TestVerdictEnum.Fail
		return result
	}

	//Передаём файлы теста в тестовый контейнер
	if err := conteinermanager.SendProjectToImage("Solutions/Gits/"+containerTestName, containerTestName, false); err != nil {
		fmt.Printf("Ошибка передачи проекта в контейнер: %v\n", err)
		result.TestingVerdict = classes.TestVerdictEnum.Fail
		return result
	}

	switch attempt.ProgrammingLanguageName {
	case "python":
		//Отправляет скрипт перенаправления запросов в хаб селениум грида
		if err := conteinermanager.ReplaceTestURLInPythonContainer(containerTestName, attempt.VariableWithURL, "http://"+containerSiteName+":80"); err != nil { //"TEST_URL"
			fmt.Printf("Ошибка замены переменной "+attempt.VariableWithURL+": %v\n", err)
			result.TestingVerdict = classes.TestVerdictEnum.Fail
			return result
		}

		if err := conteinermanager.SendSitePythonCustomize(settings.ChooseImageFilePath(attempt.ProgrammingLanguageName), "sitecustomize.py", containerTestName); err != nil {
			fmt.Println("Ошибка:", err)
			result.TestingVerdict = classes.TestVerdictEnum.Fail
			return result
		}
	case "java":
		//Отправляет скрипт перенаправления запросов в хаб селениум грида
		if err := conteinermanager.ReplaceTestURLInJavaContainer(containerTestName, attempt.VariableWithURL, "http://"+containerSiteName+":80"); err != nil { //"TEST_URL"
			fmt.Printf("Ошибка замены переменной "+attempt.VariableWithURL+": %v\n", err)
			result.TestingVerdict = classes.TestVerdictEnum.Fail
			return result
		}
		if err := conteinermanager.SendSiteJavaCustomize(settings.ChooseImageFilePath(attempt.ProgrammingLanguageName), "ChromeDriver.java", containerTestName); err != nil {
			fmt.Println("Ошибка:", err)
			result.TestingVerdict = classes.TestVerdictEnum.Fail
			return result
		}
	}

	//Загрузка сайта из гита
	if err := commands.DownloadFromGit(attempt.GitSiteURL, "main", "", "Sites/Gits/"+containerTestName, ""); err != nil {
		fmt.Printf("Ошибка загрузки сайта с гита: %v\n", err)
		result.TestingVerdict = classes.TestVerdictEnum.Fail
		return result
	}

	// Папка потока - генерируется с именем тестового контейнера
	if res, err := ExecutionSolutionOnSites("Sites/Gits/"+containerTestName+"/Sites", "Sites/Gits/"+containerTestName+"/StudentResults",
		"Sites/Gits/"+containerTestName+"/Results", containerTestName, containerSiteName, attempt.ProgrammingLanguageName); err != nil {
		fmt.Printf("Ошибка при запуске автотестов в контейнере: %v\n", err)
		return res
	} else {
		result = res
	}

	return result
}

func ExecutionSolutionOnSites(siteFolder, resultsFolder, correctResultsFolder, containerTest, containerSite, programmingLanguageName string) (classes.AllTestsInChecker, error) {
	var checkerResult classes.AllTestsInChecker

	entries, err := os.ReadDir(siteFolder)
	if err != nil {
		msg := fmt.Sprintf("Не удалось прочитать каталог %s: %v\n", siteFolder, err)
		checkerResult.Comment = msg
		checkerResult.TestingVerdict = classes.TestVerdictEnum.Fail
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
		msg := "Не найдено подпапок с числовыми именами."
		checkerResult.Comment = msg
		checkerResult.TestingVerdict = classes.TestVerdictEnum.Fail
		return checkerResult, fmt.Errorf("%s", msg)
	}

	for _, index := range dirs {
		path := dirMap[index]

		if err := conteinermanager.SendProjectToImage(path, containerSite, true); err != nil {
			msg := fmt.Sprintf("Ошибка при отправке %s: %v\n", path, err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		//Запускает тесты в контейнере
		switch programmingLanguageName {
		case "python":
			_, err := conteinermanager.RunPythonTestsContainer(containerTest)
			if err != nil {
				msg := fmt.Sprintf("Ошибка запуска Python-тестов: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}
			//Загрузка результатов с контейнера
			if err := conteinermanager.CopyResultsFromPythonContainer(containerTest, resultsFolder); err != nil {
				msg := fmt.Sprintf("Ошибка загрузки результатов с контейнера: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}
		case "java":
			_, err := conteinermanager.RunJavaTestsContainer(containerTest)
			if err != nil {
				msg := fmt.Sprintf("Ошибка выполнения Java-тестов: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}

			if err := conteinermanager.CopyResultsFromJavaContainer(containerTest, resultsFolder); err != nil {
				msg := fmt.Sprintf("Ошибка загрузки результатов с контейнера: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}
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
			msg := fmt.Sprintf("Ошибка сравнения: %v", err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		checkerResult.AllTests = append(checkerResult.AllTests, result)

		//Удаляем файлы сайта в сайтовом контейнере
		if err := conteinermanager.RemoveProjectFromContainer(containerSite, true); err != nil {
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
		} else {
			checkerResult.TestingVerdict = classes.TestVerdictEnum.Ok
		}
	}

	return checkerResult, nil
}

// ---- Отчистка ----
func ClearAllContainers(containerTestName, containerSiteName string) {

	//Удаляем файлы теста в контейнере
	if err := conteinermanager.RemoveProjectFromContainer(containerTestName, false); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта в контейнере
	if err := conteinermanager.RemoveProjectFromContainer(containerSiteName, true); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта и решения с хоста
	if err := commands.ClearHostFolder("Sites/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка удаления файлов сайта: %v\n", err)
	}
	if err := commands.ClearHostFolder("Solutions/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка удаления файлов решения: %v\n", err)
	}

	conteinermanager.RemoveTestContainer(containerTestName)
}
