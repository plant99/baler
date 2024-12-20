package baler

import (
	"errors"
	"fmt"
)

type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeIO
	ErrorTypeConfig
	ErrorTypeInternal
)

type BalerError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *BalerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s\n%v", e.Message, e.Err)
	}
	return e.Message
}

func (e *BalerError) Unwrap() error {
	return e.Err
}

// check if constructors can return BalerError
func NewValidationError(message string, err error) *BalerError {
	return &BalerError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

func NewIOError(message string, err error) *BalerError {
	return &BalerError{
		Type:    ErrorTypeIO,
		Message: message,
		Err:     err,
	}
}

func NewConfigError(message string, err error) *BalerError {
	return &BalerError{
		Type:    ErrorTypeConfig,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *BalerError {
	return &BalerError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

func IsBalerError(err error) (*BalerError, bool) {
	var balerErr *BalerError
	if err == nil {
		return nil, false
	}
	ok := errors.As(err, &balerErr)
	return balerErr, ok
}
