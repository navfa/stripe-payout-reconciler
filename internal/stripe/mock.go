package stripe

import (
	"context"
	"time"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

// MockClient is a test double for Client. Callers set the function fields
// to control return values.
type MockClient struct {
	FetchPayoutFn             func(ctx context.Context, payoutID string) (model.Payout, error)
	ListBalanceTransactionsFn func(ctx context.Context, payoutID string) ([]model.Record, error)
	ListPayoutsFn             func(ctx context.Context, from, to time.Time) ([]model.Payout, error)
}

func (m *MockClient) FetchPayout(ctx context.Context, payoutID string) (model.Payout, error) {
	return m.FetchPayoutFn(ctx, payoutID)
}

func (m *MockClient) ListBalanceTransactions(ctx context.Context, payoutID string) ([]model.Record, error) {
	return m.ListBalanceTransactionsFn(ctx, payoutID)
}

func (m *MockClient) ListPayouts(ctx context.Context, from, to time.Time) ([]model.Payout, error) {
	return m.ListPayoutsFn(ctx, from, to)
}

var _ Client = (*MockClient)(nil)
