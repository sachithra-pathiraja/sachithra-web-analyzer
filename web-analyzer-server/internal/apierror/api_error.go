package apierror

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}
