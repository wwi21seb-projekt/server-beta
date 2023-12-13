package errors

// ServerBetaError can be used to create custom errors
type ServerBetaError struct {
	Message string
	Code    string
}

// Custom error constants
var (
	INVALID_REQ_BODY = &ServerBetaError{
		Message: "The request body is invalid.",
		Code:    "ERR-001",
	}
	INVALID_EMAIL = &ServerBetaError{
		Message: "The email is invalid.",
		Code:    "ERR-002",
	}
	INVALID_USERNAME = &ServerBetaError{
		Message: "The username is invalid.",
		Code:    "ERR-003",
	}
	INVALID_NICKNAME = &ServerBetaError{
		Message: "The nickname is invalid.",
		Code:    "ERR-004",
	}
	INVALID_PASSWORD = &ServerBetaError{
		Message: "The password is invalid.",
		Code:    "ERR-005",
	}
	DATABASE_ERROR = &ServerBetaError{
		Message: "Database communication failed.",
		Code:    "ERR-006",
	}
	SERVER_ERROR = &ServerBetaError{
		Message: "Server failed.",
		Code:    "ERR-007",
	}
	EMAIL_TAKEN = &ServerBetaError{
		Message: "Email address is already taken.",
		Code:    "ERR-008",
	}
	USERNAME_TAKEN = &ServerBetaError{
		Message: "Username is already taken.",
		Code:    "ERR-009",
	}
	EMAIL_NOT_SENT = &ServerBetaError{
		Message: "Email could not be sent.",
		Code:    "ERR-010",
	}
	USER_NOT_FOUND = &ServerBetaError{
		Message: "User was not found.",
		Code:    "ERR-011",
	}
	USER_NOT_VERIFIED = &ServerBetaError{
		Message: "User is not verified.",
		Code:    "ERR-012",
	}
	INCORRECT_PASSWORD = &ServerBetaError{
		Message: "Password is incorrect.",
		Code:    "ERR-013",
	}
	VERIFICATION_TOKEN_NOT_FOUND = &ServerBetaError{
		Message: "Verification token was not found.",
		Code:    "ERR-014",
	}
	VERIFICATION_TOKEN_EXPIRED = &ServerBetaError{
		Message: "Verification token is expired. New token was sent.",
		Code:    "ERR-015",
	}
	INVALID_URL_PARAMETER = &ServerBetaError{
		Message: "Could not read needed parameters from url.",
		Code:    "ERR-016",
	}
)
