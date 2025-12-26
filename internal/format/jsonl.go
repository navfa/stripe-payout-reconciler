package format

import (
	"encoding/json"
	"io"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

// jsonlFormatter writes one JSON object per line (JSON Lines format).
type jsonlFormatter struct{}

// Format writes each record as a compact JSON object on its own line.
// An empty slice produces no output.
func (f *jsonlFormatter) Format(w io.Writer, records []model.Record) error {
	encoder := json.NewEncoder(w)
	for _, record := range records {
		if err := encoder.Encode(toJSONRecord(record)); err != nil {
			return err
		}
	}
	return nil
}
