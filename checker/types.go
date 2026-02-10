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
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var suites TestSuites

	// Попытка 1: распарсить как <testsuites>
	err1 := xml.Unmarshal(byteValue, &suites)
	if err1 == nil && len(suites.TestSuites) > 0 {
		// OK
	} else {
		// Попытка 2: распарсить как одиночный <testsuite>
		var single TestSuite
		err2 := xml.Unmarshal(byteValue, &single)
		if err2 == nil && (single.Name != "" || len(single.TestCases) > 0) {
			suites = TestSuites{TestSuites: []TestSuite{single}}
		} else {
			// Вернём оба возможных сообщения об ошибке для диагностики
			if err1 != nil && err2 != nil {
				return fmt.Errorf("ошибка парсинга XML (testsuites): %v; (testsuite): %v", err1, err2)
			}
			if err1 != nil {
				return fmt.Errorf("ошибка парсинга XML (testsuites): %v", err1)
			}
			return fmt.Errorf("не удалось распознать структуру XML")
		}
	}

	// Формируем результат в map[string]string
	results := make(map[string]string)
	for _, suite := range suites.TestSuites {
		for _, tc := range suite.TestCases {
			key := tc.Name
			// считаем, что тест упал, если есть узел <failure> или текст в failure не пуст
			if tc.Failure != nil {
				results[key] = "fail"
			} else {
				results[key] = "check"
			}
		}
	}

	// Сохраняем JSON
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
