package errors

type ServerBetaError struct {
	Message    string
	HTTPStatus int // HTTP status codes
}

func New(message string, httpStatus int) *ServerBetaError {
	return &ServerBetaError{
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func (e ServerBetaError) Error() string {
	return e.Message
}

func (e ServerBetaError) HTTPStatusCode() int {
	return e.HTTPStatus
}
