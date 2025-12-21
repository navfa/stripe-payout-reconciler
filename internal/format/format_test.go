package format_test

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
	"github.com/paco/stripe-payout-reconciler/internal/format"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		formatName  string
		wantErr     bool
		errContains string
	}{
		{name: "csv returns formatter", formatName: "csv", wantErr: false},
		{name: "json is not yet implemented", formatName: "json", wantErr: true, errContains: "not yet implemented"},
		{name: "jsonl is not yet implemented", formatName: "jsonl", wantErr: true, errContains: "not yet implemented"},
		{name: "unknown format returns error", formatName: "xml", wantErr: true, errContains: "unknown format"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			formatter, err := format.New(testCase.formatName)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("New() returned nil error, want error")
				}
				if formatter != nil {
					t.Errorf("New() returned non-nil formatter, want nil")
				}

				var inputErr *apperrors.InvalidInputError
				if !errors.As(err, &inputErr) {
					t.Errorf("New() error type = %T, want *errors.InvalidInputError", err)
				}

				if !strings.Contains(err.Error(), testCase.errContains) {
					t.Errorf("New() error = %q, want it to contain %q", err.Error(), testCase.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("New() returned unexpected error: %v", err)
				}
				if formatter == nil {
					t.Fatal("New() returned nil formatter, want non-nil")
				}
			}
		})
	}
}
