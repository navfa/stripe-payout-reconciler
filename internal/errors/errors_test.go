package errors_test

import (
	"fmt"
	"testing"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "nil error returns zero",
			err:      nil,
			wantCode: 0,
		},
		{
			name:     "InvalidInputError returns exit code 1",
			err:      apperrors.NewInvalidInputError("bad input"),
			wantCode: apperrors.ExitInvalidInput,
		},
		{
			name:     "AuthError returns exit code 2",
			err:      apperrors.NewAuthError("unauthorized"),
			wantCode: apperrors.ExitAuth,
		},
		{
			name:     "NotFoundError returns exit code 3",
			err:      apperrors.NewNotFoundError("not found", "po_123"),
			wantCode: apperrors.ExitNotFound,
		},
		{
			name:     "RateLimitError returns exit code 4",
			err:      apperrors.NewRateLimitError("slow down"),
			wantCode: apperrors.ExitRateLimit,
		},
		{
			name:     "unknown error returns exit code 99",
			err:      fmt.Errorf("something unexpected"),
			wantCode: apperrors.ExitInternal,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			gotCode := apperrors.ExitCode(testCase.err)
			if gotCode != testCase.wantCode {
				t.Errorf("ExitCode() = %d, want %d", gotCode, testCase.wantCode)
			}
		})
	}
}

func TestInvalidInputError(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := apperrors.NewInvalidInputError("payout ID must start with po_")
		if err.Error() != "payout ID must start with po_" {
			t.Errorf("Error() = %q, want %q", err.Error(), "payout ID must start with po_")
		}
		if err.Unwrap() != nil {
			t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
		}
		if err.UserMessage() != "payout ID must start with po_" {
			t.Errorf("UserMessage() = %q, want %q", err.UserMessage(), "payout ID must start with po_")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		cause := fmt.Errorf("parse error")
		err := apperrors.WrapInvalidInputError("invalid date format", cause)
		if err.Error() != "invalid date format: parse error" {
			t.Errorf("Error() = %q, want %q", err.Error(), "invalid date format: parse error")
		}
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
		}
	})
}

func TestAuthError(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := apperrors.NewAuthError("invalid API key")
		if err.Error() != "invalid API key" {
			t.Errorf("Error() = %q, want %q", err.Error(), "invalid API key")
		}
		if err.Unwrap() != nil {
			t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		cause := fmt.Errorf("401 unauthorized")
		err := apperrors.WrapAuthError("authentication failed", cause)
		if err.Error() != "authentication failed: 401 unauthorized" {
			t.Errorf("Error() = %q, want %q", err.Error(), "authentication failed: 401 unauthorized")
		}
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
		}
	})
}

func TestNotFoundError(t *testing.T) {
	t.Run("includes resource ID", func(t *testing.T) {
		err := apperrors.NewNotFoundError("payout not found", "po_abc123")
		if err.ResourceID != "po_abc123" {
			t.Errorf("ResourceID = %q, want %q", err.ResourceID, "po_abc123")
		}
		if err.Error() != "payout not found" {
			t.Errorf("Error() = %q, want %q", err.Error(), "payout not found")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		cause := fmt.Errorf("404 not found")
		err := apperrors.WrapNotFoundError("payout not found", "po_abc123", cause)
		if err.Error() != "payout not found: 404 not found" {
			t.Errorf("Error() = %q, want %q", err.Error(), "payout not found: 404 not found")
		}
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
		}
	})
}

func TestRateLimitError(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := apperrors.NewRateLimitError("rate limit exceeded")
		if err.Error() != "rate limit exceeded" {
			t.Errorf("Error() = %q, want %q", err.Error(), "rate limit exceeded")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		cause := fmt.Errorf("429 too many requests")
		err := apperrors.WrapRateLimitError("rate limit exceeded", cause)
		if err.Error() != "rate limit exceeded: 429 too many requests" {
			t.Errorf("Error() = %q, want %q", err.Error(), "rate limit exceeded: 429 too many requests")
		}
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
		}
	})
}
