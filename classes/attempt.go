package classes

import (
	"time"

	"github.com/google/uuid"
)

// Что должен слушать лисенер
type Attempt struct {
	Id                      uuid.UUID
	CreatedAt               time.Time
	GitStudentURL           string
	GitStudentBranch        string
	GitSiteURL              string // Получаю с отдельной БД (из TaskBank)
	GitSiteBranch           string // Получаю с отдельной БД (из TaskBank)
	VariableWithURL         string
	TaskId                  uuid.UUID
	TaskName                string
	ProgrammingLanguageName string
	TestingVerdict          TestVerdict // Как будто и не надо
}

//Что должен возвращать Executor
type AttemptResponse struct {
	Id                      uuid.UUID   `json:"id"`
	Results                 AllTestsInChecker `json:"results"`
}
