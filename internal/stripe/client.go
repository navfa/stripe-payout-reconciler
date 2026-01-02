// Package stripe provides the Stripe API client abstraction. The Client
// interface wraps stripe-go and translates API types into domain model
// types, ensuring stripe-go is never imported outside this package.
package stripe

import (
	"context"
	"time"

	"github.com/navfa/stripe-payout-reconciler/internal/model"
)

// Client fetches payout data from Stripe. Implementations must be safe
// for concurrent use.
type Client interface {
	// FetchPayout retrieves a single payout by its Stripe ID.
	FetchPayout(ctx context.Context, payoutID string) (model.Payout, error)

	// ListBalanceTransactions returns all balance transactions for a payout,
	// handling pagination internally.
	ListBalanceTransactions(ctx context.Context, payoutID string) ([]model.Record, error)

	// ListPayouts returns all payouts created within [from, to), handling
	// pagination internally. Both bounds are interpreted as exact timestamps.
	ListPayouts(ctx context.Context, from, to time.Time) ([]model.Payout, error)
}
