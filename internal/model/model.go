// Package model defines the core domain types for payout reconciliation.
//
// These types represent the data that flows through the application after
// being fetched from Stripe and before being formatted for output. They are
// deliberately decoupled from both the Stripe API types and the output
// format — the model package imports nothing from the rest of the project.
//
// Currency amounts are stored as int64 in the smallest currency unit (cents
// for USD, pence for GBP, yen for JPY). Conversion to decimal strings
// happens at the formatting boundary, not here. See ADR-006.
package model

import "time"

// RecordType classifies a balance transaction within a payout.
type RecordType string

const (
	// RecordTypeCharge represents a payment collected from a customer.
	RecordTypeCharge RecordType = "charge"

	// RecordTypeRefund represents a payment returned to a customer.
	RecordTypeRefund RecordType = "refund"

	// RecordTypeFee represents a Stripe processing fee.
	RecordTypeFee RecordType = "fee"

	// RecordTypeDispute represents a chargeback or dispute.
	RecordTypeDispute RecordType = "dispute"

	// RecordTypeAdjustment represents a balance adjustment by Stripe.
	RecordTypeAdjustment RecordType = "adjustment"

	// RecordTypeOther represents any transaction type not covered above.
	RecordTypeOther RecordType = "other"
)

// Record is a single balance transaction within a payout. Every field is
// always populated — there are no optional fields. One Record corresponds
// to one row in CSV output or one object in JSON/JSONL output.
//
// Amounts are in the smallest currency unit (e.g., cents for USD).
type Record struct {
	PayoutID      string
	TransactionID string
	Type          RecordType
	Amount        int64
	Fee           int64
	Net           int64
	Currency      string
	Created       time.Time
	Description   string
}

// Payout represents a Stripe payout — a transfer of funds from Stripe
// to the user's bank account. A payout contains many balance transactions,
// each represented as a Record.
type Payout struct {
	ID       string
	Amount   int64
	Currency string
	Created  time.Time
	Status   string
}
