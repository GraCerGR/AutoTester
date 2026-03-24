package classes

import (
	"time"

	"github.com/google/uuid"
)

// Что должен слушать лисенер
type Attempt struct {
	Id                      uuid.UUID
	CreatedAt               time.Time
	SolutionGit             GitInfo
	SiteGit                 GitInfo
	VariableWithURL         string
	ProgrammingLanguageName string
	Timeouts                Timeouts
	Threads                 int
	ShutdownCondition       ShutdownCondition
}

type GitInfo struct {
	URL    string
	Branch *string
}

type Timeouts struct {
	Execution time.Duration
	Test      time.Duration
}

type Threads struct {
	Number int
	Reuse  bool
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
