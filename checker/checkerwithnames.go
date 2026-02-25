package checker

import (
	"MainApp/classes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func CheckerWithNames(index int, resultsDir string, standartDir string) (classes.CheckerTest, error) {

	resultsFile := filepath.Join(resultsDir, "results_"+strconv.Itoa(index)+".json")
	expectedFile := filepath.Join(standartDir, "results_"+strconv.Itoa(index)+".json")

	actual, err := loadJSONMapWithNames(resultsFile)
	if err != nil {
		return classes.CheckerTest{}, fmt.Errorf("ошибка чтения фактического result-файла: %v", err)
	}

	expected, err := loadJSONMapWithNames(expectedFile)
	if err != nil {
		return classes.CheckerTest{}, fmt.Errorf("ошибка чтения эталонного result-файла: %v", err)
	}

	var cd classes.CheckerTest

	cd.Expected = formMapToKV(expected)
	cd.Actual = formMapToKV(actual)

	cd.TestingVerdict = classes.TestVerdictEnum.Empty
	cd.Comment = ""

	fmt.Println("Сравнение:", filepath.Base(resultsFile), "vs", filepath.Base(expectedFile))

	mismatch := false

	for key, expVal := range expected {
		actVal, exists := actual[key]

		if !exists {
			msg := fmt.Sprintf("Отсутствует тест '%s'", key)

			cd.TestingVerdict = classes.TestVerdictEnum.WrongLength
			cd.Comment += msg + "; "
			mismatch = true
			continue
		}

		if actVal != expVal {
			msg := fmt.Sprintf(
				"Несовпадение в тесте '%s' — ожидается '%s', получено '%s'",
				key, expVal, actVal,
			)

			cd.TestingVerdict = classes.TestVerdictEnum.WrongAnswer
			cd.Comment += msg + "; "
			mismatch = true
		}
	}

	for key := range actual {
		_, exists := expected[key]

		if !exists {
			msg := fmt.Sprintf("Лишний тест '%s'", key)

			cd.TestingVerdict = classes.TestVerdictEnum.WrongLength
			cd.Comment += msg + "; "
			mismatch = true
		}
	}

	if !mismatch {
		fmt.Println("Все тесты совпадают!")
		cd.TestingVerdict = classes.TestVerdictEnum.Ok
	}

	return cd, nil
}

func loadJSONMapWithNames(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data map[string]string

	dec := json.NewDecoder(file)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func formMapToKV(m map[string]string) []classes.KV {
	result := make([]classes.KV, 0, len(m))

	for k, v := range m {
		result = append(result, classes.KV{
			Key:   k,
			Value: v,
		})
	}

	return result
}
