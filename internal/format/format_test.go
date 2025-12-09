package format_test

import (
	"errors"
	"testing"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
	"github.com/paco/stripe-payout-reconciler/internal/format"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		formatName string
	}{
		{name: "csv is not yet implemented", formatName: "csv"},
		{name: "json is not yet implemented", formatName: "json"},
		{name: "jsonl is not yet implemented", formatName: "jsonl"},
		{name: "unknown format returns error", formatName: "xml"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			formatter, err := format.New(testCase.formatName)
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
		})
	}
}
