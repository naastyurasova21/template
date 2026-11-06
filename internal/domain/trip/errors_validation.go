package trip

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func newValidationError(field, msg string) *ValidationError {
	return &ValidationError{Field: field, Message: msg}
}
