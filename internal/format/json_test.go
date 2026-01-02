package format

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/navfa/stripe-payout-reconciler/internal/model"
)

func TestJSONFormatter_Format(t *testing.T) {
	fixedTime := time.Date(2025, 12, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		records []model.Record
		want    string
	}{
		{
			name:    "empty records produces empty array",
			records: []model.Record{},
			want:    "[]\n",
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
			want: `[
  {
    "payout_id": "po_abc123",
    "transaction_id": "txn_def456",
    "type": "charge",
    "amount": "150.00",
    "fee": "4.35",
    "net": "145.65",
    "currency": "usd",
    "created": "2025-12-15T14:30:00Z",
    "description": "Payment for invoice #1234"
  }
]
`,
		},
		{
			name: "multiple records",
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
			want: `[
  {
    "payout_id": "po_abc123",
    "transaction_id": "txn_001",
    "type": "charge",
    "amount": "100.00",
    "fee": "2.90",
    "net": "97.10",
    "currency": "usd",
    "created": "2025-12-15T14:30:00Z",
    "description": "Charge"
  },
  {
    "payout_id": "po_abc123",
    "transaction_id": "txn_002",
    "type": "refund",
    "amount": "-50.00",
    "fee": "0.00",
    "net": "-50.00",
    "currency": "usd",
    "created": "2025-12-15T15:30:00Z",
    "description": "Refund"
  }
]
`,
		},
		{
			name: "special characters in description",
			records: []model.Record{
				{
					PayoutID:      "po_abc123",
					TransactionID: "txn_special",
					Type:          model.RecordTypeCharge,
					Amount:        2500,
					Fee:           73,
					Net:           2427,
					Currency:      "eur",
					Created:       fixedTime,
					Description:   `Item "deluxe" — size L`,
				},
			},
			want: `[
  {
    "payout_id": "po_abc123",
    "transaction_id": "txn_special",
    "type": "charge",
    "amount": "25.00",
    "fee": "0.73",
    "net": "24.27",
    "currency": "eur",
    "created": "2025-12-15T14:30:00Z",
    "description": "Item \"deluxe\" — size L"
  }
]
`,
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
			want: `[
  {
    "payout_id": "po_jpy001",
    "transaction_id": "txn_jpy001",
    "type": "charge",
    "amount": "1500",
    "fee": "45",
    "net": "1455",
    "currency": "jpy",
    "created": "2025-12-15T14:30:00Z",
    "description": "JPY charge"
  }
]
`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &jsonFormatter{}

			err := formatter.Format(&buf, testCase.records)
			if err != nil {
				t.Fatalf("Format() returned unexpected error: %v", err)
			}

			got := buf.String()
			if got != testCase.want {
				t.Errorf("Format() output mismatch\ngot:\n%s\nwant:\n%s", got, testCase.want)
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(testCase.want, "\n")
				for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
					gotLine, wantLine := "", ""
					if i < len(gotLines) {
						gotLine = gotLines[i]
					}
					if i < len(wantLines) {
						wantLine = wantLines[i]
					}
					if gotLine != wantLine {
						t.Errorf("  line %d:\n    got:  %q\n    want: %q", i+1, gotLine, wantLine)
					}
				}
			}
		})
	}
}
