// Package format defines the output formatting interface and provides a
// factory function for creating formatters by name.
package format

import (
	"fmt"
	"io"

	apperrors "github.com/navfa/stripe-payout-reconciler/internal/errors"
	"github.com/navfa/stripe-payout-reconciler/internal/model"
)

// Formatter writes records to a writer in a specific output format
// (CSV, JSON, JSONL).
type Formatter interface {
	Format(writer io.Writer, records []model.Record) error
}

// New returns a Formatter for the given format name ("csv", "json", "jsonl").
// Returns an InvalidInputError for unrecognized or unimplemented formats.
func New(formatName string) (Formatter, error) {
	switch formatName {
	case "csv":
		return &csvFormatter{}, nil
	case "json":
		return &jsonFormatter{}, nil
	case "jsonl":
		return &jsonlFormatter{}, nil
	default:
		return nil, apperrors.NewInvalidInputError(
			fmt.Sprintf("unknown format %q", formatName),
		)
	}
}
