package apierror

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	URL     string `json:"URL"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code, message string, URL string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		URL:     URL,
	}
}
