package main

import (
	"errors"
	"testing"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
)

func TestValidatePayoutID(t *testing.T) {
	tests := []struct {
		name     string
		payoutID string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid payout ID",
			payoutID: "po_1ABC2DEF3GHI",
			wantErr:  false,
		},
		{
			name:     "minimal valid payout ID",
			payoutID: "po_x",
			wantErr:  false,
		},
		{
			name:     "missing prefix",
			payoutID: "ch_1ABC2DEF3GHI",
			wantErr:  true,
			errMsg:   `payout ID must start with "po_"`,
		},
		{
			name:     "empty string",
			payoutID: "",
			wantErr:  true,
			errMsg:   `payout ID must start with "po_"`,
		},
		{
			name:     "prefix only",
			payoutID: "po_",
			wantErr:  true,
			errMsg:   "payout ID is incomplete",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validatePayoutID(testCase.payoutID)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("validatePayoutID() returned nil error, want error")
				}
				var inputErr *apperrors.InvalidInputError
				if !errors.As(err, &inputErr) {
					t.Errorf("error type = %T, want *errors.InvalidInputError", err)
				}
				return
			}

			if err != nil {
				t.Errorf("validatePayoutID() returned unexpected error: %v", err)
			}
		})
	}
}

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{
			name:   "long key shows first 8 chars",
			apiKey: "sk_test_1234567890abcdef",
			want:   "sk_test_...",
		},
		{
			name:   "short key is fully redacted",
			apiKey: "short",
			want:   "***",
		},
		{
			name:   "exactly 8 chars is fully redacted",
			apiKey: "12345678",
			want:   "***",
		},
		{
			name:   "9 chars shows first 8",
			apiKey: "123456789",
			want:   "12345678...",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := redactAPIKey(testCase.apiKey)
			if got != testCase.want {
				t.Errorf("redactAPIKey(%q) = %q, want %q", testCase.apiKey, got, testCase.want)
			}
		})
	}
}
