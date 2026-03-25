package checker

import (
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
	Error     *Failure `xml:"error"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Content string `xml:",chardata"`
}

func Parsing(resultsFilePath string, args ...int) error {
	xmlFile, err := os.Open(resultsFilePath)
	if err != nil {
		return fmt.Errorf("Ошибка открытия файла: %w", err)
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		return fmt.Errorf("Ошибка чтения файла: %w", err)
	}

	var suites TestSuites

	err1 := xml.Unmarshal(byteValue, &suites)
	if err1 == nil && len(suites.TestSuites) > 0 {
	} else {
		var single TestSuite
		err2 := xml.Unmarshal(byteValue, &single)
		if err2 == nil && (single.Name != "" || len(single.TestCases) > 0) {
			suites = TestSuites{TestSuites: []TestSuite{single}}
		} else {
			if err1 != nil && err2 != nil {
				return fmt.Errorf("Ошибка парсинга XML (testsuites): %v; (testsuite): %v", err1, err2)
			}
			if err1 != nil {
				return fmt.Errorf("Ошибка парсинга XML (testsuites): %v", err1)
			}
			return fmt.Errorf("Не удалось распознать структуру XML")
		}
	}

	results := make(map[string]string)
	for _, suite := range suites.TestSuites {
		for _, tc := range suite.TestCases {
			key := tc.Name

			if tc.Failure != nil || tc.Error != nil {
				results[key] = "fail"
			} else {
				results[key] = "check"
			}
		}
	}

	outPath := changeExtToJSON(resultsFilePath, args...)
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("Ошибка создания файла %s: %w", outPath, err)
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		return fmt.Errorf("Ошибка записи JSON в %s: %w", outPath, err)
	}

	fmt.Printf("Результаты сохранены в %s\n", outPath)
	return nil
}

func ParseFolder(xmlFolder string, args ...string) error {
	files, err := os.ReadDir(xmlFolder)
	if err != nil {
		return fmt.Errorf("не удалось прочитать папку %s: %w", xmlFolder, err)
	}

	index := 1
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".xml" {
			continue
		}

		xmlPath := filepath.Join(xmlFolder, file.Name())
		if args != nil {
			if args[0] == "true" {
				if err := Parsing(xmlPath, index); err != nil {
					fmt.Printf("Ошибка при парсинге файла %s: %v\n", file.Name(), err)
				}
			} else {
				if err := Parsing(xmlPath); err != nil {
					fmt.Printf("Ошибка при парсинге файла %s: %v\n", file.Name(), err)
				}
			}
		} else {
			if err := Parsing(xmlPath); err != nil {
				fmt.Printf("Ошибка при парсинге файла %s: %v\n", file.Name(), err)
			}
		}
		index++
	}

	return nil
}

func changeExtToJSON(path string, index ...int) string {
	i := ""
	if index != nil {
		i = "_" + strconv.Itoa(index[0])
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return path + i + ".json"
	}
	return path[:len(path)-len(ext)] + i + ".json"
}
