package dsl

import (
	"fmt"

	"github.com/pkg/errors"
)

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Msg      string
	Pos      Position
	cause    error
	stackErr error // Error with stack trace
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s at %s: %v", e.Msg, e.Pos, e.cause)
	}
	return fmt.Sprintf("%s at %s", e.Msg, e.Pos)
}

// Line returns the line number where the error occurred
func (e *ParseError) Line() int {
	return e.Pos.Line
}

// Column returns the column number where the error occurred
func (e *ParseError) Column() int {
	return e.Pos.Column
}

// Unwrap returns the underlying cause of the error
func (e *ParseError) Unwrap() error {
	return e.stackErr
}

// newParseError creates a new ParseError
func newParseError(msg string, pos Position, cause error) *ParseError {
	var stackErr error
	if cause != nil {
		stackErr = errors.Wrap(cause, msg)
	} else {
		stackErr = errors.New(msg)
	}

	return &ParseError{
		Msg:      msg,
		Pos:      pos,
		cause:    cause,
		stackErr: stackErr,
	}
}