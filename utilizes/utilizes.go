package utilizes

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateUniqueFolder() string {

	//now := time.Now()
	//timeStr := now.Format("2006-01-02_15-04-05")
	
	folderName := fmt.Sprintf("%s", uuid.New())
	return folderName
}
