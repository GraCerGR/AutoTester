package classes

import (
	"time"

	"github.com/google/uuid"
)

type Attempt struct {
	Id                        uuid.UUID
	CreatedAt                time.Time
	GitStudentURL           string
	GitSiteURL              string
	VariableWithURL         string
	TaskId                   uuid.UUID
	TaskName                 string
	ProgrammingLanguageId   uuid.UUID
	ProgrammingLanguageName string
	TestingVerdict           TestVerdict
	Postmoderation            TestVerdict
}

type programmingLanguage struct {
	id   uuid.UUID
	name string
}
