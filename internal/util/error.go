// Package util provides utility functions for zeno.
package util

import (
	"fmt"
	"os"
)

// ExitError represents an error that should cause the program to exit.
type ExitError struct {
	Code    int
	Message string
	Err     error
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *ExitError) Unwrap() error {
	return e.Err
}

// NewExitError creates a new ExitError.
func NewExitError(code int, message string, err error) *ExitError {
	return &ExitError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// PrintError prints an error message to stderr.
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// PrintWarning prints a warning message to stderr.
func PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
}

// PrintDebug prints a debug message to stderr if debug mode is enabled.
func PrintDebug(debug bool, format string, args ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
