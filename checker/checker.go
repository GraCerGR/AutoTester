package checker

import (
	"MainApp/classes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	TestSuites []TestSuite `xml:"testsuite"`
}

type TestSuite struct {
	Name      string     `xml:"name,attr"`
	Errors    int        `xml:"errors,attr"`
	Failures  int        `xml:"failures,attr"`
	Skipped   int        `xml:"skipped,attr"`
	Tests     int        `xml:"tests,attr"`
	Time      float64    `xml:"time,attr"`
	Timestamp string     `xml:"timestamp,attr"`
	Hostname  string     `xml:"hostname,attr"`
	TestCases []TestCase `xml:"testcase"`
}

type TestCase struct {
	ClassName string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Time      float64  `xml:"time,attr"`
	Failure   *Failure `xml:"failure"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Content string `xml:",chardata"`
}

func Parsing(resultsFilePath string, index int) error {
	xmlFile, err := os.Open(resultsFilePath)
	if err != nil {
		return fmt.Errorf("Ошибка открытия файла:", err)
	}
	defer xmlFile.Close()

	byteValue, _ := io.ReadAll(xmlFile)

	var testSuites TestSuites
	if err := xml.Unmarshal(byteValue, &testSuites); err != nil {
		return fmt.Errorf("Ошибка парсинга XML:", err)
	}

	/* //Вывод в консоль
	for _, suite := range testSuites.TestSuites {
		fmt.Printf("Тестсьют: %s | Тестов: %d | Ошибок: %d | Провалов: %d\n",
			suite.Name, suite.Tests, suite.Errors, suite.Failures)

		for _, tc := range suite.TestCases {
			if tc.Failure != nil {
				fmt.Printf("NO %s.%s — FAIL\n", tc.ClassName, tc.Name)
				fmt.Printf("   Причина: %s\n", tc.Failure.Message)
			} else {
				fmt.Printf("YES %s.%s — OK\n", tc.ClassName, tc.Name)
			}
		}
	}
	*/

	results := make(map[string]string)

	for _, suite := range testSuites.TestSuites {
		for _, tc := range suite.TestCases {
			key := tc.Name
			if tc.Failure != nil {
				results[key] = "fail"
			} else {
				results[key] = "check"
			}
		}
	}

	outPath := changeExtToJSON(resultsFilePath, index)

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %w", outPath, err)
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		return fmt.Errorf("ошибка записи JSON в %s: %w", outPath, err)
	}

	fmt.Printf("Результаты сохранены в %s\n", outPath)
	return nil

}

func changeExtToJSON(path string, index int) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + "_" + strconv.Itoa(index) + ".json"
	}
	return path[:len(path)-len(ext)] + "_" + strconv.Itoa(index) + ".json"
}

type kv struct {
	Key   string
	Value string
}

type checkerTest struct {
	Expected       []kv
	Actual         []kv
	TestingVerdict classes.TestVerdict
	Comment        string
}

type AllTestsInChecker struct {
	AllTests       []checkerTest
	TestingVerdict classes.TestVerdict
	Comment        string
}

func Checker(index int, resultsDir string, standartDir string) (checkerTest, error) {

	resultsFile := filepath.Join(resultsDir, "results_"+strconv.Itoa(index)+".json")
	expectedFile := filepath.Join(standartDir, "results_"+strconv.Itoa(index)+".json")

	actual, err := loadJSONMap(resultsFile)
	if err != nil {
		return checkerTest{}, fmt.Errorf("ошибка чтения фактического result-файла: %v", err)
	}

	expected, err := loadJSONMap(expectedFile)
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

func loadJSONMap(path string) ([]kv, error) {
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
		// Ключ (строка)
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
