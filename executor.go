package main

import (
	"MainApp/checker"
	"MainApp/classes"
	"MainApp/commands"
	"MainApp/conteinermanager"
	"MainApp/errors"
	myredis "MainApp/redis"
	"MainApp/selenium"
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

func Executor(parentCtx context.Context, redisClient *redis.Client, attempt classes.Attempt, containerTestName, containerSiteName string) {

	ctx, cancel := context.WithTimeout(parentCtx, settings.ExecutionTimeout)
	defer cancel()

	var checkerResult classes.AllTestsInChecker

	defer func() {

		if r := recover(); r != nil {
			checkerResult = classes.AllTestsInChecker{
				TestingVerdict: classes.TestVerdictEnum.Fail,
				Comment:        fmt.Sprintf("Panic: %v", r),
			}
		}

		if ctx.Err() == context.DeadlineExceeded {
			checkerResult = classes.AllTestsInChecker{
				TestingVerdict: classes.TestVerdictEnum.Timeout,
				Comment:        "Превышено время выполнения",
			}
		}

		b, _ := json.MarshalIndent(checkerResult, "", "  ")
		fmt.Println("Результат проверки:")
		fmt.Println(string(b))

		Ending(redisClient, attempt, containerTestName, containerSiteName, checkerResult)
	}()

	checkerResult = ExecutorMain(ctx, attempt, containerTestName, containerSiteName)
}

func ExecutorMain(ctx context.Context, attempt classes.Attempt, containerTestName, containerSiteName string) classes.AllTestsInChecker {
	result := classes.AllTestsInChecker{}

	if err := ctx.Err(); err != nil {
		return errors.TimeoutResult()
	}

	fmt.Printf("Контейнеры выбраны для проверки:%v и %v\n", containerTestName, containerSiteName)

	//Загрузка решения из гита
	if err := commands.DownloadFromGit(ctx, attempt.GitStudentURL, "main", "", "Solutions/Gits/"+containerTestName, ""); err != nil {
		comment := fmt.Sprintf("Ошибка загрузки решения с гита: %v\n", err)
		return errors.FailResult(comment)
	}

	//Запускаем нужный контейнер для тестов
	if err := conteinermanager.RunTestContainer(ctx, containerTestName, settings.ChooseImageTag(attempt.ProgrammingLanguageName)); err != nil {
		comment := fmt.Sprintf("Не удалось создать test container: %v\n", err)
		return errors.FailResult(comment)
	}

	//Передаём файлы теста в тестовый контейнер
	if err := conteinermanager.SendProjectToImage(ctx, "Solutions/Gits/"+containerTestName, containerTestName, false); err != nil {
		comment := fmt.Sprintf("Ошибка передачи проекта в контейнер: %v\n", err)
		result.TestingVerdict = classes.TestVerdictEnum.Fail
		return errors.FailResult(comment)
	}

	switch attempt.ProgrammingLanguageName {
	case "python":
		//Отправляет скрипт перенаправления запросов в хаб селениум грида
		if err := conteinermanager.ReplaceTestURLInPythonContainer(ctx, containerTestName, attempt.VariableWithURL, "http://"+containerSiteName+":80"); err != nil { //"TEST_URL"
			comment := fmt.Sprintf("Ошибка замены переменной "+attempt.VariableWithURL+": %v\n", err)
			return errors.FailResult(comment)
		}

		if err := conteinermanager.SendSitePythonCustomize(ctx, settings.ChooseImageFilePath(attempt.ProgrammingLanguageName), "sitecustomize.py", containerTestName); err != nil {
			comment := fmt.Sprintf("Ошибка отправки sitecustomize.py: %v\n", err)
			return errors.FailResult(comment)
		}
	case "java":
		//Отправляет скрипт перенаправления запросов в хаб селениум грида
		if err := conteinermanager.ReplaceTestURLInJavaContainer(ctx, containerTestName, attempt.VariableWithURL, "http://"+containerSiteName+":80"); err != nil { //"TEST_URL"
			comment := fmt.Sprintf("Ошибка замены переменной "+attempt.VariableWithURL+": %v\n", err)
			return errors.FailResult(comment)
		}
		if err := conteinermanager.SendSiteJavaCustomize(ctx, settings.ChooseImageFilePath(attempt.ProgrammingLanguageName), "ChromeDriver.java", containerTestName); err != nil {
			comment := fmt.Sprintf("Ошибка отправки ChromeDriver.java: %v\n", err)
			return errors.FailResult(comment)
		}
	}

	//Загрузка сайта из гита
	if err := commands.DownloadFromGit(ctx, attempt.GitSiteURL, "main", "", "Sites/Gits/"+containerTestName, ""); err != nil {
		comment := fmt.Sprintf("Ошибка загрузки сайта с гита: %v\n", err)
		return errors.FailResult(comment)
	}

	// Папка потока - генерируется с именем тестового контейнера
	if res, err := ExecutionSolutionOnSites(ctx, "Sites/Gits/"+containerTestName+"/Sites", "Sites/Gits/"+containerTestName+"/StudentResults",
		"Sites/Gits/"+containerTestName+"/Results", containerTestName, containerSiteName, attempt.ProgrammingLanguageName); err != nil {
		fmt.Printf("Ошибка при запуске автотестов в контейнере: %v\n", err)
		return res
	} else {
		result = res
	}

	return result
}

func ExecutionSolutionOnSites(ctx context.Context, siteFolder, resultsFolder, correctResultsFolder, containerTest, containerSite, programmingLanguageName string) (classes.AllTestsInChecker, error) {
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

		if err := conteinermanager.SendProjectToImage(ctx, path, containerSite, true); err != nil {
			msg := fmt.Sprintf("Ошибка при отправке %s: %v\n", path, err)
			checkerResult.Comment = msg
			return checkerResult, err
		}

		//Запускает тесты в контейнере
		switch programmingLanguageName {
		case "python":
			_, err := conteinermanager.RunPythonTestsContainer(ctx, containerTest)
			if err != nil {
				msg := fmt.Sprintf("Ошибка запуска Python-тестов: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}
			//Загрузка результатов с контейнера
			if err := conteinermanager.CopyResultsFromPythonContainer(ctx, containerTest, resultsFolder); err != nil {
				msg := fmt.Sprintf("Ошибка загрузки результатов с контейнера: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}
		case "java":
			_, err := conteinermanager.RunJavaTestsContainer(ctx, containerTest)
			if err != nil {
				msg := fmt.Sprintf("Ошибка выполнения Java-тестов: %v\n", err)
				checkerResult.Comment = msg
				return checkerResult, err
			}

			if err := conteinermanager.CopyResultsFromJavaContainer(ctx, containerTest, resultsFolder); err != nil {
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
		if err := conteinermanager.RemoveProjectFromContainer(ctx, containerSite, true); err != nil {
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
func ClearAllContainers(ctx context.Context, containerTestName, containerSiteName string) {

	//Удаляем файлы теста в контейнере
	if err := conteinermanager.RemoveProjectFromContainer(ctx, containerTestName, false); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта в контейнере
	if err := conteinermanager.RemoveProjectFromContainer(ctx, containerSiteName, true); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта и решения с хоста
	if err := commands.ClearHostFolder("Sites/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка удаления файлов сайта: %v\n", err)
	}
	if err := commands.ClearHostFolder("Solutions/Gits/" + containerTestName); err != nil {
		fmt.Printf("Ошибка удаления файлов решения: %v\n", err)
	}

	conteinermanager.RemoveTestContainer(ctx, containerTestName)
}

func Ending(redisClient *redis.Client, attempt classes.Attempt, containerTestName, containerSiteName string, results classes.AllTestsInChecker) {
	ctx := context.Background()

	// Отправка результата проверки

	ClearAllContainers(ctx, containerTestName, containerSiteName)

	if err := selenium.KillSessionByName(ctx, "http://localhost:4444", containerTestName); err != nil {
		fmt.Printf("Ошибка удаления Selenium session: %v\n", err)
	}

	myredis.SetContainerStatus(ctx, redisClient, containerTestName, "free")
	myredis.SetContainerStatus(ctx, redisClient, containerSiteName, "free")
	fmt.Printf("Executor свободен для новых задач\n")
}
