package format

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

func TestCSVFormatter_Format(t *testing.T) {
	fixedTime := time.Date(2025, 12, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		records []model.Record
		want    string
	}{
		{
			name:    "empty records produces header only",
			records: []model.Record{},
			want: "payout_id,transaction_id,type,amount,fee,net,currency,created,description\n",
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
			want: "payout_id,transaction_id,type,amount,fee,net,currency,created,description\n" +
				"po_abc123,txn_def456,charge,150.00,4.35,145.65,usd,2025-12-15T14:30:00Z,Payment for invoice #1234\n",
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
			want: "payout_id,transaction_id,type,amount,fee,net,currency,created,description\n" +
				"po_abc123,txn_001,charge,100.00,2.90,97.10,usd,2025-12-15T14:30:00Z,Charge\n" +
				"po_abc123,txn_002,refund,-50.00,0.00,-50.00,usd,2025-12-15T15:30:00Z,Refund\n",
		},
		{
			name: "special characters in description are escaped",
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
					Description:   `Item "deluxe", size L`,
				},
			},
			want: "payout_id,transaction_id,type,amount,fee,net,currency,created,description\n" +
				"po_abc123,txn_special,charge,25.00,0.73,24.27,eur,2025-12-15T14:30:00Z,\"Item \"\"deluxe\"\", size L\"\n",
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
			want: "payout_id,transaction_id,type,amount,fee,net,currency,created,description\n" +
				"po_jpy001,txn_jpy001,charge,1500,45,1455,jpy,2025-12-15T14:30:00Z,JPY charge\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &csvFormatter{}

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
