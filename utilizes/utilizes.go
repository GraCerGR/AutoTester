package utilizes

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateStringId() string {

	//now := time.Now()
	//timeStr := now.Format("2006-01-02_15-04-05")
	
	id := fmt.Sprintf("%s", uuid.New())
	return id
}
