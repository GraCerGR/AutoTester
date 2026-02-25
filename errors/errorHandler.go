package errors

import (
	"MainApp/classes"
	"fmt"
)

func TimeoutResult() classes.AllTestsInChecker {
	fmt.Println("Превышено время выполнения")
	return classes.AllTestsInChecker{
		TestingVerdict: classes.TestVerdictEnum.Timeout,
		Comment:        "Превышено время выполнения",
	}
}

func FailResult(err string, allTests ...classes.CheckerTest) classes.AllTestsInChecker {
	fmt.Println(err)
	return classes.AllTestsInChecker{
		AllTests: allTests,
		TestingVerdict: classes.TestVerdictEnum.Fail,
		Comment:        err,
	}
}