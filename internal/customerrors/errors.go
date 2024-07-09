package customerrors

// Custom error constants
var (
	EndpointNotFound = &CustomError{
		Title:      "EndpointNotFound",
		Message:    "The endpoint was not found. Please check the endpoint and try again.",
		Code:       "ERR-000",
		HttpStatus: 404,
	}
	BadRequest = &CustomError{
		Title:      "BadRequest",
		Message:    "The request body is invalid. Please check the request body and try again.",
		Code:       "ERR-001",
		HttpStatus: 400,
	}
	UsernameTaken = &CustomError{
		Title:      "UsernameTaken",
		Message:    "The username is already taken. Please try another username.",
		Code:       "ERR-002",
		HttpStatus: 409,
	}
	EmailTaken = &CustomError{
		Title:      "EmailTaken",
		Message:    "The email is already taken. Please try another email.",
		Code:       "ERR-003",
		HttpStatus: 409,
	}
	UserNotFound = &CustomError{
		Title:      "UserNotFound",
		Message:    "The user was not found. Please check the username and try again.",
		Code:       "ERR-004",
		HttpStatus: 404,
	}
	UserNotActivated = &CustomError{
		Title:      "UserNotActivated",
		Message:    "The user is not activated. Please activate the user and try again.",
		Code:       "ERR-005",
		HttpStatus: 403,
	}
	ActivationTokenExpired = &CustomError{
		Title:      "ActivationTokenExpired",
		Message:    "The token has expired. Please check your mail for a new token and try again.",
		Code:       "ERR-006",
		HttpStatus: 401,
	}
	InvalidToken = &CustomError{
		Title:      "InvalidToken",
		Message:    "The token is invalid. Please check the token and try again.",
		Code:       "ERR-007",
		HttpStatus: 401,
	}
	InvalidCredentials = &CustomError{
		Title:      "InvalidCredentials",
		Message:    "The credentials are invalid. Please check the credentials and try again.",
		Code:       "ERR-008",
		HttpStatus: 401,
	}
	InternalServerError = &CustomError{
		Title:      "InternalServerError",
		Message:    "An internal server error occurred. Please try again later.",
		Code:       "ERR-009",
		HttpStatus: 500,
	}
	DatabaseError = &CustomError{
		Title:      "DatabaseError",
		Message:    "A database error occurred. Please try again later.",
		Code:       "ERR-010",
		HttpStatus: 500,
	}
	EmailUnreachable = &CustomError{
		Title:      "EmailUnreachable",
		Message:    "The email is unreachable. Please check the email and try again.",
		Code:       "ERR-011",
		HttpStatus: 422,
	}
	EmailNotSent = &CustomError{
		Title:      "EmailNotSent",
		Message:    "The email could not be sent. Please try again later.",
		Code:       "ERR-012",
		HttpStatus: 500,
	}
	UserAlreadyActivated = &CustomError{
		Title:      "UserAlreadyActivated",
		Message:    "The user is already activated. Please login to your account.",
		Code:       "ERR-013",
		HttpStatus: 208,
	}
	Unauthorized = &CustomError{
		Title:      "Unauthorized",
		Message:    "The request is unauthorized. Please login to your account.",
		Code:       "ERR-014",
		HttpStatus: 401,
	}
	SubscriptionNotFound = &CustomError{
		Title:      "SubscriptionNotFound",
		Message:    "The subscription was not found. Please check the username and try again.",
		Code:       "ERR-015",
		HttpStatus: 404,
	}
	SubscriptionAlreadyExists = &CustomError{
		Title:      "SubscriptionAlreadyExists",
		Message:    "The subscription already exists. Please check the username and try again.",
		Code:       "ERR-016",
		HttpStatus: 409,
	}
	SubscriptionSelfFollow = &CustomError{
		Title:      "SubscriptionSelfFollow",
		Message:    "You cannot follow yourself. Please check the username and try again.",
		Code:       "ERR-017",
		HttpStatus: 406,
	}
	UnsubscribeForbidden = &CustomError{
		Title:      "UnsubscribeForbidden",
		Message:    "You can only delete your own subscriptions.",
		Code:       "ERR-018",
		HttpStatus: 403,
	}
	DeletePostForbidden = &CustomError{
		Title:      "DeletePostForbidden",
		Message:    "You can only delete your own posts.",
		Code:       "ERR-019",
		HttpStatus: 403,
	}
	PostNotFound = &CustomError{
		Title:      "PostNotFound",
		Message:    "The post was not found. Please check the post ID and try again.",
		Code:       "ERR-020",
		HttpStatus: 404,
	}
	AlreadyLiked = &CustomError{
		Title:      "AlreadyLiked",
		Message:    "You have already liked this post.",
		Code:       "ERR-021",
		HttpStatus: 409,
	}
	NotLiked = &CustomError{
		Title:      "NotLiked",
		Message:    "You can't unlike a post you haven't liked.",
		Code:       "ERR-022",
		HttpStatus: 409,
	}
	NotificationNotFound = &CustomError{
		Title:      "NotificationNotFound",
		Message:    "The notification was not found. Please check the notification ID and try again.",
		Code:       "ERR-023",
		HttpStatus: 404,
	}
	DeleteNotificationForbidden = &CustomError{
		Title:      "DeleteNotificationForbidden",
		Message:    "You can only delete your own notifications.",
		Code:       "ERR-024",
		HttpStatus: 403,
	}
	PasswordResetTokenInvalid = &CustomError{
		Title:      "PasswordResetTokenInvalid",
		Message:    "The password reset token is invalid or has expired. Please request a new token and try again.",
		Code:       "ERR-025",
		HttpStatus: 403,
	}
	ChatAlreadyExists = &CustomError{
		Title:      "ChatAlreadyExists",
		Message:    "The chat already exists. Please check the username and try again.",
		Code:       "ERR-026",
		HttpStatus: 409,
	}
	ChatNotFound = &CustomError{
		Title:      "ChatNotFound",
		Message:    "The chat was not found. Please check the chat ID and try again.",
		Code:       "ERR-027",
		HttpStatus: 404,
	}
	ImageNotFound = &CustomError{
		Title:      "ImageNotFound",
		Message:    "The image was not found. Please check the image URL and try again.",
		Code:       "ERR-028",
		HttpStatus: 404,
	}
)
