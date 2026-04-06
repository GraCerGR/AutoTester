package classes

import (
	"time"

	"github.com/google/uuid"
)

// Что должен слушать лисенер
type Attempt struct {
	Id                      uuid.UUID         `json:"id"`
	CreatedAt               time.Time         `json:"created_at"`
	SolutionGit             GitInfo           `json:"solution_git"`
	SiteGit                 GitInfo           `json:"site_git"`
	VariableWithURL         string            `json:"variable_with_url"`
	ProgrammingLanguageName string            `json:"programming_language_name"`
	Timeouts                Timeouts          `json:"timeouts"`
	Threads                 int               `json:"threads"`
	ShutdownCondition       ShutdownCondition `json:"shutdown_condition"`
}

type GitInfo struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
}

type Timeouts struct {
	Execution string `json:"execution"`
	Test      string `json:"test"`
}

type Threads struct {
	Number int  `json:"number"`
	Reuse  bool `json:"reuse"`
}

type ShutdownCondition string

const (
	untilTheFirstError ShutdownCondition = "until_the_first_error"
	allTests           ShutdownCondition = "all_tests"
)

var ShutdownConditionEnum = struct {
	UntilTheFirstError ShutdownCondition
	AllTests           ShutdownCondition
}{
	UntilTheFirstError: untilTheFirstError,
	AllTests:           allTests,
}
