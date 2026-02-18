// Package errors provides custom error types for the application.
package errors

import (
	"errors"
	"fmt"
)

// Application error types.
var (
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrUnauthorized      = errors.New("user not authorized")
	ErrInvalidImage      = errors.New("invalid image")
	ErrImageTooLarge     = errors.New("image exceeds size limit")
	ErrImageTooBig       = errors.New("image dimensions exceed limit")
	ErrUploadFailed      = errors.New("image upload failed")
	ErrDownloadFailed    = errors.New("file download failed")
	ErrInvalidFileFormat = errors.New("invalid file format")
	ErrUserNotFound      = errors.New("user not found in authorized list")
	ErrUserAlreadyExists = errors.New("user already authorized")
	ErrInvalidUserID     = errors.New("invalid user ID format")
	ErrMissingFileID     = errors.New("no pending upload found")
	ErrCloudflareAPI     = errors.New("cloudflare API error")
)

// AppError represents an application-specific error with context.
type AppError struct {
	Type    error
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new application error.
func New(errType error, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Wrap wraps an existing error with an application error type.
func Wrap(errType error, message string, cause error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// Is reports whether err is of the given type.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
