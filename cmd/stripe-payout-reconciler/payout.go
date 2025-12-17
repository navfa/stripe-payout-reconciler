package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/paco/stripe-payout-reconciler/internal/config"
	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
	"github.com/paco/stripe-payout-reconciler/internal/model"
	stripeClient "github.com/paco/stripe-payout-reconciler/internal/stripe"
)

const payoutIDPrefix = "po_"

// newStripeClient is the constructor used by payoutRunE. Tests override this
// to inject a MockClient.
var newStripeClient = func(apiKey string) stripeClient.Client {
	return stripeClient.NewClient(apiKey)
}

// newPayoutCmd creates the "payout" subcommand, which fetches and displays
// balance transactions for a given Stripe payout.
func newPayoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "payout <payout-id>",
		Short: "Fetch and display balance transactions for a Stripe payout",
		Long: `Fetches all balance transactions associated with the given Stripe payout
and outputs them in a structured format for reconciliation.

The payout ID must start with "po_" (e.g., po_1ABC2DEF3GHI).`,
		Args: cobra.ExactArgs(1),
		RunE: payoutRunE,
	}
}

// payoutRunE validates the payout ID, resolves the API key, and fetches
// payout data from Stripe.
func payoutRunE(_ *cobra.Command, args []string) error {
	payoutID := args[0]

	if err := validatePayoutID(payoutID); err != nil {
		return err
	}

	cfg, err := config.Load(apiKeyFlag)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := newStripeClient(cfg.APIKey)

	payout, err := client.FetchPayout(ctx, payoutID)
	if err != nil {
		return err
	}

	records, err := client.ListBalanceTransactions(ctx, payoutID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Payout %s: %s %d (%s) — %d transactions\n",
		payout.ID,
		strings.ToUpper(payout.Currency),
		payout.Amount,
		payout.Status,
		len(records),
	)

	printRecords(os.Stdout, records)

	return nil
}

// printRecords writes records as tab-separated lines. This is a temporary
// output format that will be replaced by the formatter in Epic 3.
func printRecords(w io.Writer, records []model.Record) {
	for _, r := range records {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
			r.TransactionID,
			r.Type,
			r.Currency,
			r.Amount,
			r.Fee,
			r.Net,
			r.Created.Format(time.RFC3339),
			r.Description,
		)
	}
}

// validatePayoutID requires the "po_" prefix followed by at least one character.
func validatePayoutID(payoutID string) error {
	if !strings.HasPrefix(payoutID, payoutIDPrefix) {
		return apperrors.NewInvalidInputError(
			fmt.Sprintf("payout ID must start with %q, got %q", payoutIDPrefix, payoutID),
		)
	}

	if len(payoutID) <= len(payoutIDPrefix) {
		return apperrors.NewInvalidInputError("payout ID is incomplete: missing identifier after prefix")
	}

	return nil
}

// redactAPIKey returns a safe-to-log preview showing only the first 8
// characters, enough to distinguish sk_test_ from sk_live_.
func redactAPIKey(apiKey string) string {
	const visiblePrefix = 8
	if len(apiKey) <= visiblePrefix {
		return "***"
	}
	return apiKey[:visiblePrefix] + "..."
}
