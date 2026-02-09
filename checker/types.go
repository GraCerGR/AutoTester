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