package customerrors

import "fmt"

// CustomError can be used to create custom errors
type CustomError struct {
	Title      string `json:"title"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	HttpStatus int    `json:"http_status"`
}

// Error function to implement the error interface
func (e *CustomError) Error() string {
	return fmt.Sprintf("Message: %v Code: %v", e.Message, e.Code)
}

// ErrorResponse is used for testing to see if the right custom error was returned
type ErrorResponse struct {
	Error struct {
		Title      string `json:"title"`
		Message    string `json:"message"`
		Code       string `json:"code"`
		HttpStatus int    `json:"http_status"`
	} `json:"error"`
}
