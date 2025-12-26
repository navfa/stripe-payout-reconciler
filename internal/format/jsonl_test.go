package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

func TestJSONLFormatter_Format(t *testing.T) {
	fixedTime := time.Date(2025, 12, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		records []model.Record
		want    string
	}{
		{
			name:    "empty records produces no output",
			records: []model.Record{},
			want:    "",
		},
		{
			name: "single record",
			records: []model.Record{
				{
					PayoutID:      "po_abc123",
					TransactionID: "txn_def456",
					Type:          model.RecordTypeCharge,
					Amount:        15000,
					Fee:           435,
					Net:           14565,
					Currency:      "usd",
					Created:       fixedTime,
					Description:   "Payment for invoice #1234",
				},
			},
			want: `{"payout_id":"po_abc123","transaction_id":"txn_def456","type":"charge","amount":"150.00","fee":"4.35","net":"145.65","currency":"usd","created":"2025-12-15T14:30:00Z","description":"Payment for invoice #1234"}` + "\n",
		},
		{
			name: "multiple records produce one line each",
			records: []model.Record{
				{
					PayoutID:      "po_abc123",
					TransactionID: "txn_001",
					Type:          model.RecordTypeCharge,
					Amount:        10000,
					Fee:           290,
					Net:           9710,
					Currency:      "usd",
					Created:       fixedTime,
					Description:   "Charge",
				},
				{
					PayoutID:      "po_abc123",
					TransactionID: "txn_002",
					Type:          model.RecordTypeRefund,
					Amount:        -5000,
					Fee:           0,
					Net:           -5000,
					Currency:      "usd",
					Created:       fixedTime.Add(time.Hour),
					Description:   "Refund",
				},
			},
			want: `{"payout_id":"po_abc123","transaction_id":"txn_001","type":"charge","amount":"100.00","fee":"2.90","net":"97.10","currency":"usd","created":"2025-12-15T14:30:00Z","description":"Charge"}` + "\n" +
				`{"payout_id":"po_abc123","transaction_id":"txn_002","type":"refund","amount":"-50.00","fee":"0.00","net":"-50.00","currency":"usd","created":"2025-12-15T15:30:00Z","description":"Refund"}` + "\n",
		},
		{
			name: "zero-decimal currency",
			records: []model.Record{
				{
					PayoutID:      "po_jpy001",
					TransactionID: "txn_jpy001",
					Type:          model.RecordTypeCharge,
					Amount:        1500,
					Fee:           45,
					Net:           1455,
					Currency:      "jpy",
					Created:       fixedTime,
					Description:   "JPY charge",
				},
			},
			want: `{"payout_id":"po_jpy001","transaction_id":"txn_jpy001","type":"charge","amount":"1500","fee":"45","net":"1455","currency":"jpy","created":"2025-12-15T14:30:00Z","description":"JPY charge"}` + "\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &jsonlFormatter{}

			err := formatter.Format(&buf, testCase.records)
			if err != nil {
				t.Fatalf("Format() returned unexpected error: %v", err)
			}

			got := buf.String()
			if got != testCase.want {
				t.Errorf("Format() output mismatch\ngot:\n%s\nwant:\n%s", got, testCase.want)
			}
		})
	}
}

func TestJSONLFormatter_EachLineIsValidJSON(t *testing.T) {
	fixedTime := time.Date(2025, 12, 15, 14, 30, 0, 0, time.UTC)

	records := []model.Record{
		{
			PayoutID:      "po_abc123",
			TransactionID: "txn_001",
			Type:          model.RecordTypeCharge,
			Amount:        10000,
			Fee:           290,
			Net:           9710,
			Currency:      "usd",
			Created:       fixedTime,
			Description:   `Contains "quotes" and, commas`,
		},
		{
			PayoutID:      "po_abc123",
			TransactionID: "txn_002",
			Type:          model.RecordTypeRefund,
			Amount:        -5000,
			Fee:           0,
			Net:           -5000,
			Currency:      "usd",
			Created:       fixedTime.Add(time.Hour),
			Description:   "Normal refund",
		},
	}

	var buf bytes.Buffer
	formatter := &jsonlFormatter{}

	if err := formatter.Format(&buf, records); err != nil {
		t.Fatalf("Format() returned unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != len(records) {
		t.Fatalf("got %d lines, want %d", len(lines), len(records))
	}

	for i, line := range lines {
		var parsed jsonRecord
		if err := json.Unmarshal([]byte(line), &parsed); err != nil {
			t.Errorf("line %d is not valid JSON: %v\nline: %s", i+1, err, line)
		}
	}
}
