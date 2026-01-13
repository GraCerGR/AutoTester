package classes

import (
	"time"

	"github.com/google/uuid"
)

type Attempt struct {
	Id                      uuid.UUID
	CreatedAt               time.Time
	GitStudentURL           string
	GitStudentBranch        string
	GitSiteURL              string
	GitSiteBranch           string
	VariableWithURL         string
	TaskId                  uuid.UUID
	TaskName                string
	ProgrammingLanguageName string
	TestingVerdict          TestVerdict
}

type programmingLanguage struct {
	id   uuid.UUID
	name string
}
