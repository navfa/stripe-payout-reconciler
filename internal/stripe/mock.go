package stripe

import (
	"context"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

// MockClient is a test double for Client. Callers set the function fields
// to control return values.
type MockClient struct {
	FetchPayoutFn              func(ctx context.Context, payoutID string) (model.Payout, error)
	ListBalanceTransactionsFn  func(ctx context.Context, payoutID string) ([]model.Record, error)
}

func (m *MockClient) FetchPayout(ctx context.Context, payoutID string) (model.Payout, error) {
	return m.FetchPayoutFn(ctx, payoutID)
}

func (m *MockClient) ListBalanceTransactions(ctx context.Context, payoutID string) ([]model.Record, error) {
	return m.ListBalanceTransactionsFn(ctx, payoutID)
}

var _ Client = (*MockClient)(nil)
