package erring

type AppError struct {
	Code    string
	Message string
}

func (e AppError) Error() string { return e.Message }

func NewAppError(code, message string) AppError {
	return AppError{
		Code:    code,
		Message: message,
	}
}
