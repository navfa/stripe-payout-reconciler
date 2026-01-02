// Package errors defines structured error types for the stripe-payout-reconciler.
//
// Each error type maps to a specific failure category (authentication, not found,
// rate limiting, invalid input) and carries a deterministic exit code. This allows
// the CLI layer to translate domain errors into meaningful process exit codes
// without inspecting error strings.
//
// All error types implement the standard error interface, support unwrapping
// for errors.Is/errors.As chains, and provide a UserMessage method that returns
// a human-friendly string safe to print to stderr.
package errors

import (
	"errors"
	"fmt"
)

// Exit codes returned by the CLI process. Each error type maps to exactly one
// exit code, giving callers a reliable way to distinguish failure categories
// in scripts and CI pipelines.
const (
	ExitInvalidInput = 1
	ExitAuth         = 2
	ExitNotFound     = 3
	ExitRateLimit    = 4
	ExitInternal     = 99
)

// ExitCode returns the process exit code for err. Returns 0 for nil,
// ExitInternal for unrecognized error types.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var inputErr *InvalidInputError
	var authErr *AuthError
	var nfErr *NotFoundError
	var rlErr *RateLimitError

	switch {
	case errors.As(err, &inputErr):
		return ExitInvalidInput
	case errors.As(err, &authErr):
		return ExitAuth
	case errors.As(err, &nfErr):
		return ExitNotFound
	case errors.As(err, &rlErr):
		return ExitRateLimit
	default:
		return ExitInternal
	}
}

// InvalidInputError indicates that user-supplied input failed validation.
// Examples: malformed payout ID, invalid date format, missing required flags.
type InvalidInputError struct {
	Message string
	Err     error
}

func (e *InvalidInputError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *InvalidInputError) Unwrap() error    { return e.Err }
func (e *InvalidInputError) UserMessage() string { return e.Message }

// NewInvalidInputError returns an error for invalid user input.
func NewInvalidInputError(message string) *InvalidInputError {
	return &InvalidInputError{Message: message}
}

// WrapInvalidInputError returns an InvalidInputError wrapping cause.
func WrapInvalidInputError(message string, err error) *InvalidInputError {
	return &InvalidInputError{Message: message, Err: err}
}

// AuthError indicates an authentication or authorization failure with the
// Stripe API, typically a missing, invalid, or under-privileged API key.
type AuthError struct {
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AuthError) Unwrap() error    { return e.Err }
func (e *AuthError) UserMessage() string { return e.Message }

// NewAuthError returns an error for authentication failures.
func NewAuthError(message string) *AuthError {
	return &AuthError{Message: message}
}

// WrapAuthError returns an AuthError wrapping cause.
func WrapAuthError(message string, err error) *AuthError {
	return &AuthError{Message: message, Err: err}
}

// NotFoundError indicates that a requested resource does not exist in Stripe.
// ResourceID holds the identifier that was looked up (e.g., a payout ID).
type NotFoundError struct {
	Message    string
	ResourceID string
	Err        error
}

func (e *NotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *NotFoundError) Unwrap() error    { return e.Err }
func (e *NotFoundError) UserMessage() string { return e.Message }

// NewNotFoundError returns an error for a missing Stripe resource.
func NewNotFoundError(message string, resourceID string) *NotFoundError {
	return &NotFoundError{Message: message, ResourceID: resourceID}
}

// WrapNotFoundError returns a NotFoundError wrapping cause.
func WrapNotFoundError(message string, resourceID string, err error) *NotFoundError {
	return &NotFoundError{Message: message, ResourceID: resourceID, Err: err}
}

// RateLimitError indicates that the Stripe API returned a 429 response.
// The caller should retry with exponential backoff.
type RateLimitError struct {
	Message string
	Err     error
}

func (e *RateLimitError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *RateLimitError) Unwrap() error    { return e.Err }
func (e *RateLimitError) UserMessage() string { return e.Message }

// NewRateLimitError returns an error for Stripe rate limiting.
func NewRateLimitError(message string) *RateLimitError {
	return &RateLimitError{Message: message}
}

// WrapRateLimitError returns a RateLimitError wrapping cause.
func WrapRateLimitError(message string, err error) *RateLimitError {
	return &RateLimitError{Message: message, Err: err}
}
