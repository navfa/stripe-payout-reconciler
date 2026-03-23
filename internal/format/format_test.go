package format_test

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/navfa/stripe-payout-reconciler/internal/errors"
	"github.com/navfa/stripe-payout-reconciler/internal/format"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		formatName  string
		wantErr     bool
		errContains string
	}{
		{name: "csv returns formatter", formatName: "csv", wantErr: false},
		{name: "json returns formatter", formatName: "json", wantErr: false},
		{name: "jsonl returns formatter", formatName: "jsonl", wantErr: false},
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

				if _, ok := errors.AsType[*apperrors.InvalidInputError](err); !ok {
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
