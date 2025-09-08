package domain

var (
	ErrConflict           = "data conflicts with existing data"
	ErrNotFound           = "data not found"
	ErrInvalidParam       = "invalid parameter"
	ErrInvalidQueryParams = "invalid query parameters"
	ErrInvalidBody        = "invalid body"

	ErrInternalError   = "internal server error"
	ErrUnknownError    = "unknown error"
	ErrValidationError = "validation error"
	ErrInvalidInput    = "invalid input"
)

type ValidationError struct {
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type InternalError struct {
	Message string
	Err     error
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

type InvalidInputError struct {
	Message string
}

func (e *InvalidInputError) Error() string { return e.Message }

func NewValidationError(err error) *ValidationError {
	return &ValidationError{Message: ErrValidationError, Err: err}
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{Message: message}
}

func NewInternalError(err error) *InternalError {
	return &InternalError{Message: ErrInternalError, Err: err}
}

func NewInvalidInputError(message string) *InvalidInputError {
	return &InvalidInputError{Message: message}
}
