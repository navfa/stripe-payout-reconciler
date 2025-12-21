package format

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

// csvFormatter writes records as CSV with a header row.
type csvFormatter struct{}

// Format writes records to w as CSV. The header row is always written, even
// when records is empty. Amounts are formatted using FormatAmount for correct
// decimal placement per currency. Timestamps use RFC 3339 format.
func (f *csvFormatter) Format(w io.Writer, records []model.Record) error {
	writer := csv.NewWriter(w)

	header := []string{
		"payout_id", "transaction_id", "type",
		"amount", "fee", "net",
		"currency", "created", "description",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, record := range records {
		row := []string{
			record.PayoutID,
			record.TransactionID,
			string(record.Type),
			FormatAmount(record.Amount, record.Currency),
			FormatAmount(record.Fee, record.Currency),
			FormatAmount(record.Net, record.Currency),
			record.Currency,
			record.Created.Format(time.RFC3339),
			record.Description,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
