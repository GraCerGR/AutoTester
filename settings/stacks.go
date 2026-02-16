package settings

var Stacks = []string{"python", "java"}
var StackBuildsNames = map[string]string{
	"python": "testimagepython",
	"java":   "testimagejava",
}
var StackBuildsFiles = map[string]string{
	"python": "DockerfilePython.base",
	"java":   "DockerfileJava.base",
}

var StackBuildsFilesPaths = map[string]string{
	"python": "dockerfiles/python/",
	"java":   "dockerfiles/java/",
}


func ChooseImageTag(stack string) string {
	if imageName, ok := StackBuildsNames[stack]; ok {
		return imageName
	}
	return "unknown"
}

func ChooseImageFile(stack string) string {
	if imageFile, ok := StackBuildsFiles[stack]; ok {
		return imageFile
	}
	return "unknown"
}

func ChooseImageFilePath(stack string) string {
	if imageFile, ok := StackBuildsFilesPaths[stack]; ok {
		return imageFile
	}
	return "unknown"
}
