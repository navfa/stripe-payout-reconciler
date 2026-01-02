package stripe

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stripe/stripe-go/v82"

	apperrors "github.com/navfa/stripe-payout-reconciler/internal/errors"
	"github.com/navfa/stripe-payout-reconciler/internal/model"
)

func TestMapTransactionType(t *testing.T) {
	tests := []struct {
		stripeType string
		want       model.RecordType
	}{
		{"charge", model.RecordTypeCharge},
		{"payment", model.RecordTypeCharge},
		{"refund", model.RecordTypeRefund},
		{"payment_refund", model.RecordTypeRefund},
		{"payment_failure_refund", model.RecordTypeRefund},
		{"stripe_fee", model.RecordTypeFee},
		{"stripe_fx_fee", model.RecordTypeFee},
		{"tax_fee", model.RecordTypeFee},
		{"application_fee", model.RecordTypeFee},
		{"dispute", model.RecordTypeDispute},
		{"issuing_dispute", model.RecordTypeDispute},
		{"adjustment", model.RecordTypeAdjustment},
		{"unknown", model.RecordTypeOther},
		{"payout", model.RecordTypeOther},
		{"transfer", model.RecordTypeOther},
		{"topup", model.RecordTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.stripeType, func(t *testing.T) {
			var buf bytes.Buffer
			origWriter := warnWriter
			warnWriter = &buf
			defer func() { warnWriter = origWriter }()

			got := mapTransactionType(tt.stripeType)
			if got != tt.want {
				t.Errorf("mapTransactionType(%q) = %q, want %q", tt.stripeType, got, tt.want)
			}

			if tt.want == model.RecordTypeOther {
				if !strings.Contains(buf.String(), "unknown balance transaction type") {
					t.Errorf("expected warning for unknown type %q, got %q", tt.stripeType, buf.String())
				}
			} else {
				if buf.Len() > 0 {
					t.Errorf("unexpected warning for known type %q: %q", tt.stripeType, buf.String())
				}
			}
		})
	}
}

func TestTranslateError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantNil    bool
		wantType   string // "auth", "notfound", "ratelimit", "generic"
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:     "401 returns AuthError",
			err:      &stripe.Error{HTTPStatusCode: http.StatusUnauthorized, Msg: "invalid key"},
			wantType: "auth",
		},
		{
			name:     "403 returns AuthError",
			err:      &stripe.Error{HTTPStatusCode: http.StatusForbidden, Msg: "forbidden"},
			wantType: "auth",
		},
		{
			name:     "404 returns NotFoundError",
			err:      &stripe.Error{HTTPStatusCode: http.StatusNotFound, Msg: "not found"},
			wantType: "notfound",
		},
		{
			name:     "429 returns RateLimitError",
			err:      &stripe.Error{HTTPStatusCode: http.StatusTooManyRequests, Msg: "rate limited"},
			wantType: "ratelimit",
		},
		{
			name:     "500 returns generic error",
			err:      &stripe.Error{HTTPStatusCode: http.StatusInternalServerError, Msg: "server error"},
			wantType: "generic",
		},
		{
			name:     "non-stripe error returns generic error",
			err:      fmt.Errorf("connection refused"),
			wantType: "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateError("po_test123", tt.err)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("translateError() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("translateError() = nil, want error")
			}

			switch tt.wantType {
			case "auth":
				var authErr *apperrors.AuthError
				if !errors.As(got, &authErr) {
					t.Errorf("error type = %T, want *AuthError", got)
				}
			case "notfound":
				var nfErr *apperrors.NotFoundError
				if !errors.As(got, &nfErr) {
					t.Errorf("error type = %T, want *NotFoundError", got)
				}
			case "ratelimit":
				var rlErr *apperrors.RateLimitError
				if !errors.As(got, &rlErr) {
					t.Errorf("error type = %T, want *RateLimitError", got)
				}
			case "generic":
				var authErr *apperrors.AuthError
				var nfErr *apperrors.NotFoundError
				var rlErr *apperrors.RateLimitError
				if errors.As(got, &authErr) || errors.As(got, &nfErr) || errors.As(got, &rlErr) {
					t.Errorf("error type = %T, want generic (non-app) error", got)
				}
			}
		})
	}
}
