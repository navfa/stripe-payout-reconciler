package model

import (
	"testing"
	"time"
)

func TestSummarize(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		records    []Record
		wantLen    int
		wantFirst  Summary
	}{
		{
			name:    "nil records",
			records: nil,
			wantLen: 0,
		},
		{
			name:    "empty records",
			records: []Record{},
			wantLen: 0,
		},
		{
			name: "single currency mixed types",
			records: []Record{
				{Type: RecordTypeCharge, Amount: 10000, Fee: 300, Net: 9700, Currency: "usd", Created: now},
				{Type: RecordTypeCharge, Amount: 5000, Fee: 150, Net: 4850, Currency: "usd", Created: now},
				{Type: RecordTypeRefund, Amount: -2000, Fee: 0, Net: -2000, Currency: "usd", Created: now},
				{Type: RecordTypeFee, Amount: -450, Fee: 0, Net: -450, Currency: "usd", Created: now},
			},
			wantLen: 1,
			wantFirst: Summary{
				Currency: "usd",
				ByType: []TypeSummary{
					{Type: RecordTypeCharge, Count: 2, Amount: 15000, Fee: 450, Net: 14550},
					{Type: RecordTypeRefund, Count: 1, Amount: -2000, Fee: 0, Net: -2000},
					{Type: RecordTypeFee, Count: 1, Amount: -450, Fee: 0, Net: -450},
					{Type: RecordTypeDispute, Count: 0, Amount: 0, Fee: 0, Net: 0},
					{Type: RecordTypeAdjustment, Count: 0, Amount: 0, Fee: 0, Net: 0},
					{Type: RecordTypeOther, Count: 0, Amount: 0, Fee: 0, Net: 0},
				},
				Total: TypeSummary{Count: 4, Amount: 12550, Fee: 450, Net: 12100},
			},
		},
		{
			name: "charges only",
			records: []Record{
				{Type: RecordTypeCharge, Amount: 1000, Fee: 30, Net: 970, Currency: "eur", Created: now},
			},
			wantLen: 1,
			wantFirst: Summary{
				Currency: "eur",
				ByType: []TypeSummary{
					{Type: RecordTypeCharge, Count: 1, Amount: 1000, Fee: 30, Net: 970},
					{Type: RecordTypeRefund, Count: 0},
					{Type: RecordTypeFee, Count: 0},
					{Type: RecordTypeDispute, Count: 0},
					{Type: RecordTypeAdjustment, Count: 0},
					{Type: RecordTypeOther, Count: 0},
				},
				Total: TypeSummary{Count: 1, Amount: 1000, Fee: 30, Net: 970},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Summarize(tt.records)
			if len(got) != tt.wantLen {
				t.Fatalf("Summarize() returned %d summaries, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen == 0 {
				return
			}

			s := got[0]
			if s.Currency != tt.wantFirst.Currency {
				t.Errorf("Currency = %q, want %q", s.Currency, tt.wantFirst.Currency)
			}
			if len(s.ByType) != len(tt.wantFirst.ByType) {
				t.Fatalf("ByType len = %d, want %d", len(s.ByType), len(tt.wantFirst.ByType))
			}
			for i, want := range tt.wantFirst.ByType {
				got := s.ByType[i]
				if got != want {
					t.Errorf("ByType[%d] = %+v, want %+v", i, got, want)
				}
			}
			if s.Total != tt.wantFirst.Total {
				t.Errorf("Total = %+v, want %+v", s.Total, tt.wantFirst.Total)
			}
		})
	}
}

func TestSummarizeMultiCurrency(t *testing.T) {
	now := time.Now()
	records := []Record{
		{Type: RecordTypeCharge, Amount: 10000, Fee: 300, Net: 9700, Currency: "usd", Created: now},
		{Type: RecordTypeCharge, Amount: 5000, Fee: 150, Net: 4850, Currency: "eur", Created: now},
		{Type: RecordTypeRefund, Amount: -1000, Fee: 0, Net: -1000, Currency: "usd", Created: now},
	}

	got := Summarize(records)
	if len(got) != 2 {
		t.Fatalf("Summarize() returned %d summaries, want 2", len(got))
	}

	if got[0].Currency != "usd" {
		t.Errorf("first summary currency = %q, want %q", got[0].Currency, "usd")
	}
	if got[1].Currency != "eur" {
		t.Errorf("second summary currency = %q, want %q", got[1].Currency, "eur")
	}

	if got[0].Total.Count != 2 {
		t.Errorf("usd total count = %d, want 2", got[0].Total.Count)
	}
	if got[1].Total.Count != 1 {
		t.Errorf("eur total count = %d, want 1", got[1].Total.Count)
	}
}

func TestSummarizeCanonicalOrder(t *testing.T) {
	now := time.Now()
	// Insert in reverse order — output should still follow canonical order.
	records := []Record{
		{Type: RecordTypeOther, Amount: 100, Currency: "usd", Created: now},
		{Type: RecordTypeCharge, Amount: 200, Currency: "usd", Created: now},
	}

	got := Summarize(records)
	if len(got) != 1 {
		t.Fatalf("Summarize() returned %d summaries, want 1", len(got))
	}

	wantOrder := []RecordType{
		RecordTypeCharge, RecordTypeRefund, RecordTypeFee,
		RecordTypeDispute, RecordTypeAdjustment, RecordTypeOther,
	}
	for i, want := range wantOrder {
		if got[0].ByType[i].Type != want {
			t.Errorf("ByType[%d].Type = %q, want %q", i, got[0].ByType[i].Type, want)
		}
	}
}
