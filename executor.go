package main

import (
	"MainApp/classes"
	"MainApp/commands"
	"MainApp/conteinermanager"
	myerrors "MainApp/errors"
	myredis "MainApp/redis"
	"MainApp/selenium"
	"MainApp/settings"
	"MainApp/utilizes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

func Executor(parentCtx context.Context, redisClient *redis.Client, attempt classes.Attempt, containerTestName string, containersSiteName []string) {

	ctx, cancel := context.WithTimeout(parentCtx, attempt.Timeouts.Execution)
	defer cancel()

	var checkerResult classes.AllTestsInChecker

	defer func() {

		if r := recover(); r != nil {
			checkerResult = myerrors.FailResult(fmt.Sprintf("Panic: %v", r))
		}

		if ctx.Err() == context.DeadlineExceeded {
			checkerResult = myerrors.TimeoutResult()
		}

		b, _ := json.MarshalIndent(checkerResult, "", "  ")
		fmt.Println("Результат проверки:")
		fmt.Println(string(b))

		Ending(redisClient, containerTestName, containersSiteName, checkerResult, attempt)
	}()

	checkerResult = ExecutorMain(ctx, attempt, containerTestName, containersSiteName)
}

func ExecutorMain(ctx context.Context, attempt classes.Attempt, containerTestName string, containersSiteName []string) classes.AllTestsInChecker {
	result := classes.AllTestsInChecker{}

	if err := ctx.Err(); err != nil {
		return myerrors.TimeoutResult()
	}

	fmt.Printf("Контейнеры выбраны для проверки:%v и %v\n", containerTestName, containersSiteName)

	//Создание папкок для хранения
	if err := createFolders(attempt); err != nil {
		comment := fmt.Sprintf("Ошибка создания папок: %v\n", err)
		return myerrors.FailResult(comment)
	}

	//Загрузка решения из гита
	if err := commands.DownloadFromGit(ctx, attempt.SolutionGit.URL, attempt.SolutionGit.Branch, "", settings.FolderSolution+containerTestName, ""); err != nil {
		comment := fmt.Sprintf("Ошибка загрузки решения с гита: %v\n", err)
		return myerrors.FailResult(comment)
	}

	//Запускаем нужный контейнер для тестов
	if err := conteinermanager.RunTestContainer(ctx, containerTestName, settings.ChooseImageTag(attempt.ProgrammingLanguageName)); err != nil {
		comment := fmt.Sprintf("Не удалось создать test container: %v\n", err)
		return myerrors.FailResult(comment)
	}

	//Передаём файлы теста в тестовый контейнер
	if err := conteinermanager.SendProjectToImage(ctx, settings.FolderSolution+containerTestName, containerTestName, false); err != nil {
		comment := fmt.Sprintf("Ошибка передачи проекта в контейнер: %v\n", err)
		return myerrors.FailResult(comment)
	}

	switch attempt.ProgrammingLanguageName {
	case "python":
		if err := conteinermanager.AddTestURLImportToPythonFiles(ctx, containerTestName, attempt.VariableWithURL); err != nil {
			comment := fmt.Sprintf("Ошибка добавления import "+attempt.VariableWithURL+": %v\n", err)
			return myerrors.FailResult(comment)
		}

		if err := conteinermanager.CommentPythonVariableInContainer(ctx, containerTestName, attempt.VariableWithURL); err != nil {
			comment := fmt.Sprintf("Ошибка комментирования переменной "+attempt.VariableWithURL+": %v\n", err)
			return myerrors.FailResult(comment)
		}

		if err := conteinermanager.SendSitePythonCustomize(ctx, settings.ChooseImageFilePath(attempt.ProgrammingLanguageName), "sitecustomize.py", containerTestName); err != nil {
			comment := fmt.Sprintf("Ошибка отправки sitecustomize.py: %v\n", err)
			return myerrors.FailResult(comment)
		}
	case "java":
		if err := conteinermanager.AddTestURLImportToJavaFiles(ctx, containerTestName, attempt.VariableWithURL); err != nil {
			comment := fmt.Sprintf("Ошибка добавления import"+attempt.VariableWithURL+": %v\n", err)
			return myerrors.FailResult(comment)
		}

		if err := conteinermanager.CommentJavaVariableInContainer(ctx, containerTestName, attempt.VariableWithURL); err != nil {
			comment := fmt.Sprintf("Ошибка комментирования переменной %s: %v\n", attempt.VariableWithURL, err)
			return myerrors.FailResult(comment)
		}

		if err := conteinermanager.SendSiteJavaCustomize(ctx, settings.ChooseImageFilePath(attempt.ProgrammingLanguageName), "ChromeDriver.java", containerTestName); err != nil {
			comment := fmt.Sprintf("Ошибка отправки ChromeDriver.java: %v\n", err)
			return myerrors.FailResult(comment)
		}
	}

	//Загрузка сайта из гита
	if err := commands.DownloadFromGit(ctx, attempt.SiteGit.URL, attempt.SiteGit.Branch, "", settings.FolderSite+containerTestName, ""); err != nil {
		comment := fmt.Sprintf("Ошибка загрузки сайта с гита: %v\n", err)
		return myerrors.FailResult(comment)
	}

	if checkerAllResults, err := ExecutionSolutionOnSites(ctx, settings.FolderSite+containerTestName+"/Sites", settings.FolderSite+containerTestName+"/StudentResults",
		settings.FolderSite+containerTestName+"/Results", containerTestName, containersSiteName, attempt); err != nil {
		fmt.Printf("Ошибка при запуске автотестов в контейнере: %v\n", err)
		return checkerAllResults
	} else {
		result = checkerAllResults
	}

	return result
}

func ExecutionSolutionOnSites(ctx context.Context, siteFolder, resultsFolder, correctResultsFolder, containerTest string, containersSite []string, attempt classes.Attempt) (classes.AllTestsInChecker, error) {
	var checkerAllResults classes.AllTestsInChecker

	dirs, dirMap, err := utilizes.CountingFolder(siteFolder)
	if err != nil {
		msg := fmt.Sprintf("Ошибка подсчёта папок в %s: %v\n", siteFolder, err)
		checkerAllResults.Comment = msg
		checkerAllResults.TestingVerdict = classes.TestVerdictEnum.Fail
		return checkerAllResults, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	checkerAllResults.AllTests = make([]classes.CheckerTest, len(dirs))

	threads := attempt.Threads

	if threads > len(dirs) {
		threads = len(dirs)
	}

	var stop bool = false

	for t := 0; t < threads; t++ {

		wg.Add(1)

		go func(threadIndex int) {
			defer wg.Done()

			for i := threadIndex; i < len(dirs); i += threads {

				if attempt.ShutdownCondition == classes.ShutdownConditionEnum.UntilTheFirstError {
					mu.Lock()
					if stop {
						mu.Unlock()
						return
					}
					mu.Unlock()
				}

				index := dirs[i]
				path := dirMap[index]
				var checkerOneResult classes.CheckerTest
				checkerOneResult.Id = index
				checkerAllResults.AllTests[i].Id = checkerOneResult.Id

				if err := conteinermanager.SendProjectToImage(ctx, path, containersSite[threadIndex], true); err != nil {
					msg := fmt.Sprintf("Ошибка при отправке %s: %v\n", path, err)

					checkerOneResult.Comment = msg
					checkerOneResult.TestingVerdict = classes.TestVerdictEnum.Fail

					mu.Lock()
					checkerAllResults.Comment = msg
					checkerAllResults.TestingVerdict = checkerOneResult.TestingVerdict
					checkerAllResults.AllTests[i] = checkerOneResult
					stop = true
					mu.Unlock()

					return
				}

				if launchResult, err := LaunchTestsInConteiner(ctx, containerTest, containersSite[threadIndex], resultsFolder, attempt, index); err != nil {
					fmt.Printf("Ошибка запуска тестов в контейнере: %v\n", err)

					mu.Lock()
					checkerAllResults.Comment = launchResult.Comment
					checkerAllResults.TestingVerdict = launchResult.TestingVerdict
					checkerAllResults.AllTests[i] = launchResult
					stop = true
					mu.Unlock()

					_ = conteinermanager.RemoveProjectFromContainer(ctx, containersSite[threadIndex], true)
					continue
				}

				if checkerTest, err := Checker(ctx, index, resultsFolder, correctResultsFolder); err != nil {
					fmt.Printf("Ошибка при проверке результатов: %v\n", err)

					checkerOneResult.Comment = checkerTest.Comment
					checkerOneResult.TestingVerdict = checkerTest.TestingVerdict

					mu.Lock()
					checkerAllResults.Comment = checkerTest.Comment
					checkerAllResults.TestingVerdict = checkerTest.TestingVerdict
					checkerAllResults.AllTests[i] = checkerOneResult
					stop = true
					mu.Unlock()

					_ = conteinermanager.RemoveProjectFromContainer(ctx, containersSite[threadIndex], true)
					continue

				} else {
					checkerOneResult.Comment = checkerTest.Comment
					checkerOneResult.TestingVerdict = checkerTest.TestingVerdict
					checkerOneResult.Expected = checkerTest.Expected
					checkerOneResult.Actual = checkerTest.Actual
				}

				if err := conteinermanager.RemoveProjectFromContainer(ctx, containersSite[threadIndex], true); err != nil {
					msg := fmt.Sprintf("Ошибка при очистке контейнера: %v\n", err)

					checkerOneResult.Comment = msg
					checkerOneResult.TestingVerdict = classes.TestVerdictEnum.Fail

					mu.Lock()
					checkerAllResults.Comment = checkerOneResult.Comment
					checkerAllResults.TestingVerdict = checkerOneResult.TestingVerdict
					checkerAllResults.AllTests[i] = checkerOneResult
					stop = true
					mu.Unlock()
					continue
				}

				mu.Lock()
				checkerAllResults.AllTests[i] = checkerOneResult
				if checkerOneResult.TestingVerdict != classes.TestVerdictEnum.Ok {
					checkerAllResults.Comment = checkerOneResult.Comment
					checkerAllResults.TestingVerdict = checkerOneResult.TestingVerdict
					stop = true
				}
				mu.Unlock()
			}
		}(t)
	}
	wg.Wait()

	if checkerAllResults.TestingVerdict == classes.TestVerdictEnum.Null {
		checkerAllResults.TestingVerdict = classes.TestVerdictEnum.Ok
	}

	var newAllTests []classes.CheckerTest
	for i := range checkerAllResults.AllTests {
		if checkerAllResults.AllTests[i].Id != 0 || checkerAllResults.AllTests[i].TestingVerdict != classes.TestVerdictEnum.Null {
			newAllTests = append(newAllTests, checkerAllResults.AllTests[i])
		}
	}

	checkerAllResults.AllTests = newAllTests

	return checkerAllResults, nil
}

func LaunchTestsInConteiner(parentCtx context.Context, containerTest, containerSite, resultsFolder string, attempt classes.Attempt, index int) (classes.CheckerTest, error) {
	var launchResult classes.CheckerTest
	launchResult.Id = index

	ctx, cancel := context.WithTimeout(parentCtx, attempt.Timeouts.Test)
	defer cancel()

	logFilePath := fmt.Sprintf("%s/%d.log", settings.FolderLog+attempt.Id.String(), index)

	switch attempt.ProgrammingLanguageName {
	case "python":

		_, err := conteinermanager.RunPythonTestsContainer(ctx, containerTest, "http://"+containerSite+":80", index, logFilePath)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				launchResult.TestingVerdict = classes.TestVerdictEnum.Timeout
				launchResult.Comment = "Превышено время выполнения тестов"
				return launchResult, err
			}

			msg := fmt.Sprintf("Ошибка запуска Python-тестов: %v\n", err)
			launchResult.TestingVerdict = classes.TestVerdictEnum.FailLaunchTests
			launchResult.Comment = msg
			return launchResult, err
		}

		if err := conteinermanager.CopyResultsFromPythonContainer(ctx, containerTest, resultsFolder, index); err != nil {
			msg := fmt.Sprintf("Ошибка загрузки результатов с контейнера: %v\n", err)
			launchResult.TestingVerdict = classes.TestVerdictEnum.FailLoadTestResults
			launchResult.Comment = msg
			return launchResult, err
		}
	case "java":

		_, err := conteinermanager.RunJavaTestsContainer(ctx, containerTest, "http://"+containerSite+":80", index, logFilePath)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				launchResult.TestingVerdict = classes.TestVerdictEnum.Timeout
				launchResult.Comment = "Превышено время выполнения тестов"
				return launchResult, err
			}

			msg := fmt.Sprintf("Ошибка выполнения Java-тестов: %v\n", err)
			launchResult.TestingVerdict = classes.TestVerdictEnum.FailLaunchTests
			launchResult.Comment = msg
			return launchResult, err
		}

		if err := conteinermanager.CopyResultsFromJavaContainer(ctx, containerTest, resultsFolder, index); err != nil {
			msg := fmt.Sprintf("Ошибка загрузки результатов с контейнера: %v\n", err)
			launchResult.TestingVerdict = classes.TestVerdictEnum.FailLoadTestResults
			launchResult.Comment = msg
			return launchResult, err
		}
	}

	return launchResult, nil
}

// ---- Отчистка ----
func ClearAllContainers(ctx context.Context, containerTestName string, containerSiteName []string) {

	//Удаляем файлы теста в контейнере
	if err := conteinermanager.RemoveProjectFromContainer(ctx, containerTestName, false); err != nil {
		fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
	}

	//Удаляем файлы сайта в контейнере
	for index := range containerSiteName {
		if err := conteinermanager.RemoveProjectFromContainer(ctx, containerSiteName[index], true); err != nil {
			fmt.Printf("Ошибка при очистке контейнера: %v\n", err)
		}
	}

	//Удаляем файлы сайта и решения с хоста
	if err := commands.ClearHostFolder(settings.FolderSite + containerTestName); err != nil {
		fmt.Printf("Ошибка удаления файлов сайта: %v\n", err)
	}
	if err := commands.ClearHostFolder(settings.FolderSolution + containerTestName); err != nil {
		fmt.Printf("Ошибка удаления файлов решения: %v\n", err)
	}

	conteinermanager.RemoveTestContainer(ctx, containerTestName)
}

func Ending(redisClient *redis.Client, containerTestName string, containerSiteName []string, results classes.AllTestsInChecker, attempt classes.Attempt) {
	ctx := context.Background()

	// Отправка результата проверки (пока просто сохранение в файл)
	b, _ := json.MarshalIndent(results, "", "  ")
	if err := os.WriteFile(settings.FolderLog+attempt.Id.String()+"/result.json", b, 0644); err != nil {
		fmt.Printf("Ошибка создания файла результата: %v\n", err)
	}

	ClearAllContainers(ctx, containerTestName, containerSiteName)

	if err := selenium.KillSessionByName(ctx, settings.HubURL, containerTestName); err != nil {
		fmt.Printf("Ошибка удаления Selenium session: %v\n", err)
	}

	myredis.SetContainerStatus(ctx, redisClient, containerTestName, "free")

	for index := range containerSiteName {
		myredis.SetContainerStatus(ctx, redisClient, containerSiteName[index], "free")
	}

	fmt.Printf("Executor свободен для новых задач\n")
}

func createFolders(attempt classes.Attempt) error {
	if err := commands.CreateFolder(settings.FolderLog + attempt.Id.String()); err != nil {
		return err
	}
	if err := commands.CreateFolder(settings.FolderSolution); err != nil {
		return err
	}
	if err := commands.CreateFolder(settings.FolderSite); err != nil {
		return err
	}
	return nil
}
