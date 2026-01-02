package format

import (
	"encoding/json"
	"io"
	"time"

	"github.com/navfa/stripe-payout-reconciler/internal/model"
)

// jsonRecord is the serialization shape shared by the JSON and JSONL formatters.
// Amounts are strings (not float64) to avoid IEEE 754 rounding in financial data.
type jsonRecord struct {
	PayoutID      string `json:"payout_id"`
	TransactionID string `json:"transaction_id"`
	Type          string `json:"type"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
	Net           string `json:"net"`
	Currency      string `json:"currency"`
	Created       string `json:"created"`
	Description   string `json:"description"`
}

// toJSONRecord converts a domain Record to its JSON serialization shape.
func toJSONRecord(record model.Record) jsonRecord {
	return jsonRecord{
		PayoutID:      record.PayoutID,
		TransactionID: record.TransactionID,
		Type:          string(record.Type),
		Amount:        FormatAmount(record.Amount, record.Currency),
		Fee:           FormatAmount(record.Fee, record.Currency),
		Net:           FormatAmount(record.Net, record.Currency),
		Currency:      record.Currency,
		Created:       record.Created.Format(time.RFC3339),
		Description:   record.Description,
	}
}

// jsonFormatter writes records as an indented JSON array.
type jsonFormatter struct{}

// Format writes records to w as a JSON array with 2-space indentation.
// An empty slice produces "[]" followed by a newline.
func (f *jsonFormatter) Format(w io.Writer, records []model.Record) error {
	output := make([]jsonRecord, len(records))
	for i, record := range records {
		output[i] = toJSONRecord(record)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
