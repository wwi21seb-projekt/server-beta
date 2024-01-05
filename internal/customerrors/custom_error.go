package customerrors

import "fmt"

// CustomError can be used to create custom customerrors
type CustomError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error function to implement the error interface
func (e *CustomError) Error() string {
	return fmt.Sprintf("%v:%v: syntax error", e.Message, e.Code)
}

// ErrorResponse is used for testing to see if the right custom error was returned
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Custom error constants
var (
	BadRequest = &CustomError{
		Message: "The request body is invalid. Please check the request body and try again.",
		Code:    "ERR-001",
	}
	UsernameTaken = &CustomError{
		Message: "The username is already taken. Please try another username.",
		Code:    "ERR-002",
	}
	EmailTaken = &CustomError{
		Message: "The email is already taken. Please try another email.",
		Code:    "ERR-003",
	}
	UserNotFound = &CustomError{
		Message: "The user was not found. Please check the username and try again.",
		Code:    "ERR-004",
	}
	UserNotActivated = &CustomError{
		Message: "The user is not activated. Please activate the user and try again.",
		Code:    "ERR-005",
	}
	ActivationTokenExpired = &CustomError{
		Message: "The token has expired. Please check your mail for a new token and try again.",
		Code:    "ERR-006",
	}
	InvalidToken = &CustomError{
		Message: "The token is invalid. Please check the token and try again.",
		Code:    "ERR-007",
	}
	InvalidCredentials = &CustomError{
		Message: "The credentials are invalid. Please check the credentials and try again.",
		Code:    "ERR-008",
	}
	InternalServerError = &CustomError{
		Message: "An internal server error occurred. Please try again later.",
		Code:    "ERR-009",
	}
	DatabaseError = &CustomError{
		Message: "A database error occurred. Please try again later.",
		Code:    "ERR-010",
	}
	EmailUnreachable = &CustomError{
		Message: "The email is unreachable. Please check the email and try again.",
		Code:    "ERR-011",
	}
	EmailNotSent = &CustomError{
		Message: "The email could not be sent. Please try again later.",
		Code:    "ERR-012",
	}
	UserAlreadyActivated = &CustomError{
		Message: "The user is already activated. Please login to your account.",
		Code:    "ERR-013",
	}
	PreliminaryUserUnauthorized = &CustomError{
		Message: "The user is not authorized. Please login to your account.",
		Code:    "ERR-014",
	}
	PreliminaryFileNotFound = &CustomError{
		Message: "The file was not found. Please check the file and try again.",
		Code:    "ERR-015",
	}
	PreliminaryFileTooLarge = &CustomError{
		Message: "The file is too large. Please compress file.",
		Code:    "ERR-016",
	}
)
