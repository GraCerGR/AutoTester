package classes

type TestVerdict string

const (
	verdictOk          TestVerdict = "OK"
	verdictWrongAnswer TestVerdict = "WRONG_ANSWER"
	verdictFail        TestVerdict = "FAIL"
	verdictEmpty       TestVerdict = ""
	verdictWrongLength TestVerdict = "WRONG_LENGTH"
	verdictTimeout     TestVerdict = "TIMEOUT"
)

var TestVerdictEnum = struct {
	WrongAnswer TestVerdict
	Ok          TestVerdict
	Fail        TestVerdict
	Empty       TestVerdict
	WrongLength TestVerdict
	Timeout     TestVerdict
}{
	Ok:          verdictOk,
	Fail:        verdictFail,
	Empty:       verdictEmpty,
	WrongAnswer: verdictWrongAnswer,
	WrongLength: verdictWrongLength,
	Timeout:     verdictTimeout,
}
