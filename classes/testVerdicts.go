package classes

type TestVerdict string

const (
	verdictOk              TestVerdict = "OK"
	verdictWrongAnswer     TestVerdict = "WRONG_ANSWER"
	verdictFail            TestVerdict = "FAIL"
	verdictEmpty           TestVerdict = ""
	verdictWrongLength     TestVerdict = "WRONG_LENGTH"
	verdictTimeout         TestVerdict = "TIMEOUT"
	verdictFailLaunchTests TestVerdict = "FAIL_LAUNCH_TESTS"
	verdictCheckerError    TestVerdict = "CHECKER_ERROR"
	verdictNull            TestVerdict = ""
)

var TestVerdictEnum = struct {
	WrongAnswer     TestVerdict
	Ok              TestVerdict
	Fail            TestVerdict
	Empty           TestVerdict
	WrongLength     TestVerdict
	Timeout         TestVerdict
	FailLaunchTests TestVerdict
	CheckerError    TestVerdict
	Null            TestVerdict
}{
	Ok:              verdictOk,
	Fail:            verdictFail,
	Empty:           verdictEmpty,
	WrongAnswer:     verdictWrongAnswer,
	WrongLength:     verdictWrongLength,
	Timeout:         verdictTimeout,
	FailLaunchTests: verdictFailLaunchTests,
	CheckerError:    verdictCheckerError,
	Null:            verdictNull,
}

type KV struct {
	Key   string
	Value string
}

type CheckerTest struct {
	Id             int
	Expected       []KV
	Actual         []KV
	TestingVerdict TestVerdict
	Comment        string
}

type AllTestsInChecker struct {
	AllTests       []CheckerTest
	TestingVerdict TestVerdict
	Comment        string
}
