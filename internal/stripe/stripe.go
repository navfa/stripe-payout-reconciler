package stripe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/stripe/stripe-go/v82"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
	"github.com/paco/stripe-payout-reconciler/internal/model"
)

// apiClient implements Client using the stripe-go SDK with a per-instance
// backend, avoiding global state so multiple clients can coexist.
type apiClient struct {
	sc *stripe.Client
}

// NewClient returns a Client that calls the Stripe API with the given key.
// Safe for concurrent use.
func NewClient(apiKey string) Client {
	backendCfg := &stripe.BackendConfig{
		MaxNetworkRetries: stripe.Int64(5),
	}
	backend := stripe.GetBackendWithConfig(stripe.APIBackend, backendCfg)

	sc := stripe.NewClient(apiKey, stripe.WithBackends(&stripe.Backends{
		API:         &stripe.UsageBackend{B: backend, Usage: []string{"stripe_client_new"}},
		Connect:     &stripe.UsageBackend{B: stripe.GetBackend(stripe.ConnectBackend), Usage: []string{"stripe_client_new"}},
		Uploads:     &stripe.UsageBackend{B: stripe.GetBackend(stripe.UploadsBackend), Usage: []string{"stripe_client_new"}},
		MeterEvents: &stripe.UsageBackend{B: stripe.GetBackend(stripe.MeterEventsBackend), Usage: []string{"stripe_client_new"}},
	}))

	return &apiClient{sc: sc}
}

func (c *apiClient) FetchPayout(ctx context.Context, payoutID string) (model.Payout, error) {
	p, err := c.sc.V1Payouts.Retrieve(ctx, payoutID, nil)
	if err != nil {
		return model.Payout{}, translateError(payoutID, err)
	}

	return model.Payout{
		ID:       p.ID,
		Amount:   p.Amount,
		Currency: string(p.Currency),
		Created:  time.Unix(p.Created, 0),
		Status:   string(p.Status),
	}, nil
}

func (c *apiClient) ListBalanceTransactions(ctx context.Context, payoutID string) ([]model.Record, error) {
	params := &stripe.BalanceTransactionListParams{}
	params.Payout = stripe.String(payoutID)
	params.Limit = stripe.Int64(100)

	var records []model.Record
	for bt, err := range c.sc.V1BalanceTransactions.List(ctx, params) {
		if err != nil {
			return nil, translateError(payoutID, err)
		}
		records = append(records, model.Record{
			PayoutID:      payoutID,
			TransactionID: bt.ID,
			Type:          mapTransactionType(string(bt.Type)),
			Amount:        bt.Amount,
			Fee:           bt.Fee,
			Net:           bt.Net,
			Currency:      string(bt.Currency),
			Created:       time.Unix(bt.Created, 0),
			Description:   bt.Description,
		})
	}

	return records, nil
}

// translateError converts a stripe-go error into the application's structured
// error types by inspecting stripe.Error fields (Code, HTTPStatusCode).
func translateError(resourceID string, err error) error {
	if err == nil {
		return nil
	}

	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		switch stripeErr.HTTPStatusCode {
		case 401, 403:
			return apperrors.WrapAuthError("authentication failed: check your Stripe API key", err)
		case 404:
			return apperrors.WrapNotFoundError(
				fmt.Sprintf("resource not found: %s", resourceID), resourceID, err,
			)
		case 429:
			return apperrors.WrapRateLimitError("rate limited by Stripe API", err)
		default:
			return fmt.Errorf("stripe API error (HTTP %d) for %s: %w", stripeErr.HTTPStatusCode, resourceID, err)
		}
	}

	return fmt.Errorf("stripe request failed for %s: %w", resourceID, err)
}

// mapTransactionType converts a Stripe balance transaction type string
// to the application's RecordType.
func mapTransactionType(t string) model.RecordType {
	switch t {
	case "charge", "payment":
		return model.RecordTypeCharge
	case "refund", "payment_refund", "payment_failure_refund":
		return model.RecordTypeRefund
	case "stripe_fee", "stripe_fx_fee", "tax_fee", "application_fee":
		return model.RecordTypeFee
	case "dispute", "issuing_dispute":
		return model.RecordTypeDispute
	case "adjustment":
		return model.RecordTypeAdjustment
	default:
		fmt.Fprintf(os.Stderr, "warning: unknown balance transaction type %q, mapped to %q\n", t, model.RecordTypeOther)
		return model.RecordTypeOther
	}
}

var _ Client = (*apiClient)(nil)
