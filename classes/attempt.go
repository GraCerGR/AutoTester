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
	GitSiteURL              string
	GitSiteBranch           string
	VariableWithURL         string
	ProgrammingLanguageName string
}
