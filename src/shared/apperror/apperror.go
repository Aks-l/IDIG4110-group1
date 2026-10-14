package apperror

const (
	CodeEmailAlreadyUsed = "email_already_used"
	CodeInvalidInput     = "invalid_input"
	CodeInternal         = "internal_error"
	CodeNoContent        = "no_content"
	CodeUnauthorized     = "unauthorized_request"
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func CreateNewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func InternalServerError() *Error {
	return CreateNewError(
		CodeInternal,
		"internal server error",
	)
}
