package service

type ErrorBadRequest struct {
	errMsg string
}

func (e *ErrorBadRequest) Error() string {
	return e.errMsg
}

type ErrorCorruptedDatabaseResult struct {
	errMsg string
}

func (e *ErrorCorruptedDatabaseResult) Error() string {
	return e.errMsg
}
