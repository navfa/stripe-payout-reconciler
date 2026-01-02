package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	apperrors "github.com/navfa/stripe-payout-reconciler/internal/errors"
	"github.com/navfa/stripe-payout-reconciler/internal/model"
	stripeClient "github.com/navfa/stripe-payout-reconciler/internal/stripe"
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

func TestParseDateFlag(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantY   int
		wantM   time.Month
		wantD   int
		wantErr bool
	}{
		{name: "valid date", value: "2024-01-15", wantY: 2024, wantM: time.January, wantD: 15},
		{name: "end of year", value: "2024-12-31", wantY: 2024, wantM: time.December, wantD: 31},
		{name: "invalid month", value: "2024-13-01", wantErr: true},
		{name: "not a date", value: "not-a-date", wantErr: true},
		{name: "wrong format", value: "2024-1-5", wantErr: true},
		{name: "empty string", value: "", wantErr: true},
		{name: "ISO datetime rejected", value: "2024-01-15T00:00:00Z", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := parseDateFlag(testCase.value)

			if testCase.wantErr {
				if err == nil {
					t.Fatalf("parseDateFlag(%q) returned nil error, want error", testCase.value)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseDateFlag(%q) returned unexpected error: %v", testCase.value, err)
			}

			if got.Year() != testCase.wantY || got.Month() != testCase.wantM || got.Day() != testCase.wantD {
				t.Errorf("parseDateFlag(%q) = %v, want %d-%02d-%02d",
					testCase.value, got, testCase.wantY, testCase.wantM, testCase.wantD)
			}

			if got.Location() != time.UTC {
				t.Errorf("parseDateFlag(%q) location = %v, want UTC", testCase.value, got.Location())
			}
		})
	}
}

func TestValidatePayoutFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		from        string
		to          string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid single payout",
			args: []string{"po_abc123"},
		},
		{
			name: "valid date range",
			from: "2024-01-01",
			to:   "2024-01-31",
		},
		{
			name: "same from and to date",
			from: "2024-01-15",
			to:   "2024-01-15",
		},
		{
			name:        "payout ID with --from is mutually exclusive",
			args:        []string{"po_abc123"},
			from:        "2024-01-01",
			to:          "2024-01-31",
			wantErr:     true,
			errContains: "mutually exclusive",
		},
		{
			name:        "payout ID with --to is mutually exclusive",
			args:        []string{"po_abc123"},
			to:          "2024-01-31",
			wantErr:     true,
			errContains: "mutually exclusive",
		},
		{
			name:        "--from without --to",
			from:        "2024-01-01",
			wantErr:     true,
			errContains: "both --from and --to",
		},
		{
			name:        "--to without --from",
			to:          "2024-01-31",
			wantErr:     true,
			errContains: "both --from and --to",
		},
		{
			name:        "no arguments and no flags",
			wantErr:     true,
			errContains: "provide a payout ID",
		},
		{
			name:        "reversed date range",
			from:        "2024-01-31",
			to:          "2024-01-01",
			wantErr:     true,
			errContains: "must not be before",
		},
		{
			name:        "invalid --from date",
			from:        "not-a-date",
			to:          "2024-01-31",
			wantErr:     true,
			errContains: "invalid --from",
		},
		{
			name:        "invalid --to date",
			from:        "2024-01-01",
			to:          "not-a-date",
			wantErr:     true,
			errContains: "invalid --to",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			origFrom, origTo := fromFlag, toFlag
			defer func() { fromFlag, toFlag = origFrom, origTo }()
			fromFlag = testCase.from
			toFlag = testCase.to

			err := validatePayoutFlags(testCase.args)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("validatePayoutFlags() returned nil error, want error")
				}
				var inputErr *apperrors.InvalidInputError
				if !errors.As(err, &inputErr) {
					t.Errorf("error type = %T, want *errors.InvalidInputError", err)
				}
				if !strings.Contains(err.Error(), testCase.errContains) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), testCase.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("validatePayoutFlags() returned unexpected error: %v", err)
			}
		})
	}
}

func TestPeriodReconciliation_Success(t *testing.T) {
	originalClient := newStripeClient
	defer func() { newStripeClient = originalClient }()

	origFormat, origFrom, origTo := formatFlag, fromFlag, toFlag
	defer func() { formatFlag, fromFlag, toFlag = origFormat, origFrom, origTo }()
	formatFlag = "csv"
	fromFlag = "2024-01-01"
	toFlag = "2024-01-31"

	mock := &stripeClient.MockClient{
		ListPayoutsFn: func(_ context.Context, _, _ time.Time) ([]model.Payout, error) {
			return []model.Payout{
				{ID: "po_001", Amount: 50000, Currency: "usd", Status: "paid"},
				{ID: "po_002", Amount: 30000, Currency: "usd", Status: "paid"},
			}, nil
		},
		ListBalanceTransactionsFn: func(_ context.Context, payoutID string) ([]model.Record, error) {
			return []model.Record{
				{
					PayoutID:      payoutID,
					TransactionID: "txn_" + payoutID,
					Type:          model.RecordTypeCharge,
					Amount:        10000,
					Fee:           290,
					Net:           9710,
					Currency:      "usd",
					Created:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
					Description:   "Payment",
				},
			}, nil
		},
	}
	newStripeClient = func(_ string) stripeClient.Client { return mock }

	t.Setenv("STRIPE_API_KEY", "sk_test_fake")

	err := payoutRunE(nil, []string{})
	if err != nil {
		t.Fatalf("payoutRunE() returned unexpected error: %v", err)
	}
}

func TestPeriodReconciliation_EmptyRange(t *testing.T) {
	originalClient := newStripeClient
	defer func() { newStripeClient = originalClient }()

	origFormat, origFrom, origTo := formatFlag, fromFlag, toFlag
	defer func() { formatFlag, fromFlag, toFlag = origFormat, origFrom, origTo }()
	formatFlag = "csv"
	fromFlag = "2024-06-01"
	toFlag = "2024-06-30"

	mock := &stripeClient.MockClient{
		ListPayoutsFn: func(_ context.Context, _, _ time.Time) ([]model.Payout, error) {
			return []model.Payout{}, nil
		},
	}
	newStripeClient = func(_ string) stripeClient.Client { return mock }

	t.Setenv("STRIPE_API_KEY", "sk_test_fake")

	err := payoutRunE(nil, []string{})
	if err != nil {
		t.Fatalf("payoutRunE() returned unexpected error: %v", err)
	}
}
