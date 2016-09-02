package commands

type FailedError struct {
}

func (err FailedError) Error() string {
	return "FAILED"
}
