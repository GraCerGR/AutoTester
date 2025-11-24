package classes

import (
	"time"

	"github.com/google/uuid"
)

var Attempt struct {
	id                        uuid.UUID
	created_at                time.Time
	git_student_url           string
	git_site_url              string
	task_id                   uuid.UUID
	task_name                 string
	programming_language_id   uuid.UUID
	programming_language_name string
	testing_verdict           TestVerdict
	postmoderation            TestVerdict
}

var ProgrammingLanguage struct {
	id   uuid.UUID
	name string
}
