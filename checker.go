package main

import (
	"MainApp/checker"
	"MainApp/classes"
	"context"
	"fmt"
	"strconv"
)

func Checker(ctx context.Context, index int, resultsFolder, correctResultsFolder string) (classes.CheckerTest, error) {
	var checkerResult classes.CheckerTest

	// ---- Checker ----

	//Вывод реузльтатов теста в json
	if err := checker.Parsing(resultsFolder + "/results_"+ strconv.Itoa(index) +".xml"); err != nil {
		msg := fmt.Sprintf("Ошибка вывода результатов в json: %v", err)
		checkerResult.Comment = msg
		checkerResult.TestingVerdict = classes.TestVerdictEnum.CheckerError
		return checkerResult, err
	}

	result, err := checker.CheckerWithNames(index, resultsFolder, correctResultsFolder)
	if err != nil {
		msg := fmt.Sprintf("Ошибка сравнения: %v", err)
		checkerResult.Comment = msg
		checkerResult.TestingVerdict = classes.TestVerdictEnum.CheckerError
		return checkerResult, err
	}

	return result, nil
}
