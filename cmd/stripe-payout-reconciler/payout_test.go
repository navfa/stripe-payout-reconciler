package main

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
	"github.com/paco/stripe-payout-reconciler/internal/model"
	stripeClient "github.com/paco/stripe-payout-reconciler/internal/stripe"
)

func TestValidatePayoutID(t *testing.T) {
	tests := []struct {
		name     string
		payoutID string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid payout ID",
			payoutID: "po_1ABC2DEF3GHI",
			wantErr:  false,
		},
		{
			name:     "minimal valid payout ID",
			payoutID: "po_x",
			wantErr:  false,
		},
		{
			name:     "missing prefix",
			payoutID: "ch_1ABC2DEF3GHI",
			wantErr:  true,
			errMsg:   `payout ID must start with "po_"`,
		},
		{
			name:     "empty string",
			payoutID: "",
			wantErr:  true,
			errMsg:   `payout ID must start with "po_"`,
		},
		{
			name:     "prefix only",
			payoutID: "po_",
			wantErr:  true,
			errMsg:   "payout ID is incomplete",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validatePayoutID(testCase.payoutID)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("validatePayoutID() returned nil error, want error")
				}
				var inputErr *apperrors.InvalidInputError
				if !errors.As(err, &inputErr) {
					t.Errorf("error type = %T, want *errors.InvalidInputError", err)
				}
				return
			}

			if err != nil {
				t.Errorf("validatePayoutID() returned unexpected error: %v", err)
			}
		})
	}
}

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{
			name:   "long key shows first 8 chars",
			apiKey: "sk_test_1234567890abcdef",
			want:   "sk_test_...",
		},
		{
			name:   "short key is fully redacted",
			apiKey: "short",
			want:   "***",
		},
		{
			name:   "exactly 8 chars is fully redacted",
			apiKey: "12345678",
			want:   "***",
		},
		{
			name:   "9 chars shows first 8",
			apiKey: "123456789",
			want:   "12345678...",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := redactAPIKey(testCase.apiKey)
			if got != testCase.want {
				t.Errorf("redactAPIKey(%q) = %q, want %q", testCase.apiKey, got, testCase.want)
			}
		})
	}
}

func TestPayoutRunE_Success(t *testing.T) {
	originalClient := newStripeClient
	defer func() { newStripeClient = originalClient }()

	originalFormat := formatFlag
	defer func() { formatFlag = originalFormat }()
	formatFlag = "csv"

	mock := &stripeClient.MockClient{
		FetchPayoutFn: func(_ context.Context, _ string) (model.Payout, error) {
			return model.Payout{
				ID:       "po_test123",
				Amount:   100000,
				Currency: "usd",
				Created:  time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				Status:   "paid",
			}, nil
		},
		ListBalanceTransactionsFn: func(_ context.Context, _ string) ([]model.Record, error) {
			return []model.Record{
				{
					PayoutID:      "po_test123",
					TransactionID: "txn_1",
					Type:          model.RecordTypeCharge,
					Amount:        50000,
					Fee:           1500,
					Net:           48500,
					Currency:      "usd",
					Created:       time.Date(2025, 11, 28, 14, 0, 0, 0, time.UTC),
					Description:   "Payment for invoice #1",
				},
			}, nil
		},
	}
	newStripeClient = func(_ string) stripeClient.Client { return mock }

	t.Setenv("STRIPE_API_KEY", "sk_test_fake")

	err := payoutRunE(nil, []string{"po_test123"})
	if err != nil {
		t.Fatalf("payoutRunE() returned unexpected error: %v", err)
	}
}

func TestPayoutRunE_NotFound(t *testing.T) {
	original := newStripeClient
	defer func() { newStripeClient = original }()

	mock := &stripeClient.MockClient{
		FetchPayoutFn: func(_ context.Context, _ string) (model.Payout, error) {
			return model.Payout{}, apperrors.NewNotFoundError("resource not found: po_missing", "po_missing")
		},
		ListBalanceTransactionsFn: func(_ context.Context, _ string) ([]model.Record, error) {
			return nil, nil
		},
	}
	newStripeClient = func(_ string) stripeClient.Client { return mock }

	t.Setenv("STRIPE_API_KEY", "sk_test_fake")

	err := payoutRunE(nil, []string{"po_missing"})
	if err == nil {
		t.Fatal("payoutRunE() returned nil error, want NotFoundError")
	}

	var nfErr *apperrors.NotFoundError
	if !errors.As(err, &nfErr) {
		t.Errorf("error type = %T, want *NotFoundError", err)
	}
}

func TestPayoutRunE_AuthError(t *testing.T) {
	original := newStripeClient
	defer func() { newStripeClient = original }()

	mock := &stripeClient.MockClient{
		FetchPayoutFn: func(_ context.Context, _ string) (model.Payout, error) {
			return model.Payout{}, apperrors.NewAuthError("authentication failed: check your Stripe API key")
		},
		ListBalanceTransactionsFn: func(_ context.Context, _ string) ([]model.Record, error) {
			return nil, nil
		},
	}
	newStripeClient = func(_ string) stripeClient.Client { return mock }

	t.Setenv("STRIPE_API_KEY", "sk_test_fake")

	err := payoutRunE(nil, []string{"po_test123"})
	if err == nil {
		t.Fatal("payoutRunE() returned nil error, want AuthError")
	}

	var authErr *apperrors.AuthError
	if !errors.As(err, &authErr) {
		t.Errorf("error type = %T, want *AuthError", err)
	}
}
