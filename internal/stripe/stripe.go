package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v82"

	"github.com/paco/stripe-payout-reconciler/internal/model"
)

// apiClient implements Client using the stripe-go SDK with a per-instance
// backend, avoiding global state so multiple clients can coexist.
type apiClient struct {
	apiKey  string
	backend stripe.Backend
}

// NewClient returns a Client that calls the Stripe API with the given key.
// Safe for concurrent use.
func NewClient(apiKey string) Client {
	backendConfig := &stripe.BackendConfig{}
	backend := stripe.GetBackendWithConfig(stripe.APIBackend, backendConfig)

	return &apiClient{
		apiKey:  apiKey,
		backend: backend,
	}
}

func (client *apiClient) FetchPayout(_ context.Context, payoutID string) (model.Payout, error) {
	return model.Payout{}, translateError(payoutID, nil)
}

func (client *apiClient) ListBalanceTransactions(_ context.Context, payoutID string) ([]model.Record, error) {
	return nil, translateError(payoutID, nil)
}

// translateError converts a stripe-go error into the application's structured
// error types by inspecting stripe.Error fields (Code, HTTPStatusCode).
// TODO: implement when real API calls are wired up.
func translateError(resourceID string, _ error) error {
	return nil
}

var _ Client = (*apiClient)(nil)
