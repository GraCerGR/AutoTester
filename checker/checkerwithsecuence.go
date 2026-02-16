package checker

import (
	"MainApp/classes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func CheckerWithSequence(index int, resultsDir string, standartDir string) (checkerTest, error) {

	resultsFile := filepath.Join(resultsDir, "results_"+strconv.Itoa(index)+".json")
	expectedFile := filepath.Join(standartDir, "results_"+strconv.Itoa(index)+".json")

	actual, err := loadJSONMapWithSequence(resultsFile)
	if err != nil {
		return checkerTest{}, fmt.Errorf("ошибка чтения фактического result-файла: %v", err)
	}

	expected, err := loadJSONMapWithSequence(expectedFile)
	if err != nil {
		return checkerTest{}, fmt.Errorf("ошибка чтения эталонного result-файла: %v", err)
	}

	var cd checkerTest
	cd.Expected = expected
	cd.Actual = actual
	cd.TestingVerdict = classes.TestVerdictEnum.Empty
	cd.Comment = ""

	fmt.Println("Сравнение:", filepath.Base(resultsFile), "vs", filepath.Base(expectedFile))

	mismatch := false

	minLen := len(expected)
	if len(actual) < minLen {
		minLen = len(actual)
	}

	for i := 0; i < minLen; i++ {
		expVal := expected[i].Value
		actVal := actual[i].Value
		if actVal != expVal {
			msg := fmt.Sprintf("Несовпадение в позиции %d — ожидается '%s', получено '%s'", i, expVal, actVal)

			cd.TestingVerdict = classes.TestVerdictEnum.WrongAnswer
			cd.Comment = msg
			mismatch = true
		}
	}

	if len(actual) > len(expected) {
		for i := len(expected); i < len(actual); i++ {
			msg := fmt.Sprintf("Лишний параметр в фактическом файле на позиции %d: ключ '%s' значение %s",
				i, actual[i].Key, actual[i].Value)

			cd.TestingVerdict = classes.TestVerdictEnum.WrongLength
			cd.Comment = msg
			mismatch = true
		}
	}

	if len(expected) > len(actual) {
		for i := len(actual); i < len(expected); i++ {
			msg := fmt.Sprintf("Отсутствует параметр в фактическом файле на позиции %d: ожидаемый ключ '%s' значение %s",
				i, expected[i].Key, expected[i].Value)

			cd.TestingVerdict = classes.TestVerdictEnum.WrongLength
			cd.Comment = msg
			mismatch = true
		}
	}

	if !mismatch {
		fmt.Println("Все тесты совпадают!")
		cd.TestingVerdict = classes.TestVerdictEnum.Ok
	}

	return cd, nil
}

func loadJSONMapWithSequence(path string) ([]kv, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("ожидается JSON-объект в файле %s", path)
	}

	var pairs []kv

	for dec.More() {

		tk, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := tk.(string)
		if !ok {
			return nil, fmt.Errorf("ожидается строковый ключ в файле %s", path)
		}

		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}

		var sval string
		if err := json.Unmarshal(raw, &sval); err == nil {
			pairs = append(pairs, kv{Key: key, Value: sval})
		} else {
			pairs = append(pairs, kv{Key: key, Value: string(raw)})
		}
	}

	if _, err := dec.Token(); err != nil {
		return nil, err
	}

	return pairs, nil
}
