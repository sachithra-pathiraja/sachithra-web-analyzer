package apierror

// Error represents a client-side API error.
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	URL     string `json:"URL,omitempty"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// New creates a new Error value.
func New(code, message string, url string) *Error {
	return &Error{Code: code, Message: message, URL: url}
}

// Client-side error codes
const (
	ErrRequestCreation = "request_creation_failed"
	ErrRequestFailed   = "request_failed"
	ErrReadFailed      = "read_failed"
	ErrInvalidResponse = "invalid_response"
	ErrServerError     = "server_error"
	ErrUnmarshalFailed = "unmarshal_failed"
)
