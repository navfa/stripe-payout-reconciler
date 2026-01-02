package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/navfa/stripe-payout-reconciler/internal/config"
	apperrors "github.com/navfa/stripe-payout-reconciler/internal/errors"
	"github.com/navfa/stripe-payout-reconciler/internal/format"
	"github.com/navfa/stripe-payout-reconciler/internal/model"
	stripeClient "github.com/navfa/stripe-payout-reconciler/internal/stripe"
)

const payoutIDPrefix = "po_"

// newStripeClient is the constructor used by payoutRunE. Tests override this
// to inject a MockClient.
var newStripeClient = func(apiKey string) stripeClient.Client {
	return stripeClient.NewClient(apiKey)
}

var (
	formatFlag string
	fromFlag   string
	toFlag     string
)

// newPayoutCmd creates the "payout" subcommand, which fetches and displays
// balance transactions for a given Stripe payout or all payouts in a date range.
func newPayoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payout [payout-id]",
		Short: "Fetch and display balance transactions for Stripe payouts",
		Long: `Fetches balance transactions and outputs them for reconciliation.

Two modes of operation:

  Single payout:  stripe-payout-reconciler payout po_1ABC2DEF3GHI
  Date range:     stripe-payout-reconciler payout --from 2024-01-01 --to 2024-01-31

The payout ID and --from/--to flags are mutually exclusive.
Dates are interpreted as UTC (YYYY-MM-DD format).`,
		Example: `  # Reconcile a single payout to CSV (default)
  stripe-payout-reconciler payout po_1ABC2DEF3GHI

  # Export as JSON
  stripe-payout-reconciler payout po_1ABC2DEF3GHI --format json

  # Reconcile all payouts in January 2024
  stripe-payout-reconciler payout --from 2024-01-01 --to 2024-01-31

  # Pipe JSONL to jq for filtering
  stripe-payout-reconciler payout po_1ABC2DEF3GHI --format jsonl | jq 'select(.type == "charge")'`,
		Args: cobra.MaximumNArgs(1),
		RunE: payoutRunE,
	}

	cmd.Flags().StringVar(&formatFlag, "format", "csv",
		`output format: "csv", "json", or "jsonl"`,
	)
	cmd.Flags().StringVar(&fromFlag, "from", "",
		"start date (inclusive, UTC, YYYY-MM-DD)",
	)
	cmd.Flags().StringVar(&toFlag, "to", "",
		"end date (inclusive, UTC, YYYY-MM-DD)",
	)

	return cmd
}

// payoutRunE validates inputs and delegates to single-payout or period mode.
func payoutRunE(_ *cobra.Command, args []string) error {
	if err := validatePayoutFlags(args); err != nil {
		return err
	}

	cfg, err := config.Load(apiKeyFlag)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := newStripeClient(cfg.APIKey)

	if fromFlag != "" {
		return runPeriodReconciliation(ctx, client)
	}
	return runSinglePayout(ctx, client, args[0])
}

// runSinglePayout fetches and formats transactions for a single payout ID.
func runSinglePayout(ctx context.Context, client stripeClient.Client, payoutID string) error {
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

	formatter, err := format.New(formatFlag)
	if err != nil {
		return err
	}

	return formatter.Format(os.Stdout, records)
}

// runPeriodReconciliation fetches all payouts in the date range and formats
// their combined transactions.
func runPeriodReconciliation(ctx context.Context, client stripeClient.Client) error {
	from, _ := parseDateFlag(fromFlag)
	to, _ := parseDateFlag(toFlag)
	toExclusive := to.Add(24 * time.Hour)

	payouts, err := client.ListPayouts(ctx, from, toExclusive)
	if err != nil {
		return err
	}

	if len(payouts) == 0 {
		fmt.Fprintf(os.Stderr, "No payouts found in range %s to %s\n", fromFlag, toFlag)
		return nil
	}

	var allRecords []model.Record
	for idx, payout := range payouts {
		fmt.Fprintf(os.Stderr, "Fetching payout %d of %d (%s)...\n", idx+1, len(payouts), payout.ID)
		records, err := client.ListBalanceTransactions(ctx, payout.ID)
		if err != nil {
			return err
		}
		allRecords = append(allRecords, records...)
	}

	fmt.Fprintf(os.Stderr, "%d payouts, %d transactions\n", len(payouts), len(allRecords))

	formatter, err := format.New(formatFlag)
	if err != nil {
		return err
	}

	return formatter.Format(os.Stdout, allRecords)
}

// validatePayoutFlags enforces mutual exclusivity between the positional
// payout ID and the --from/--to date range flags.
func validatePayoutFlags(args []string) error {
	hasID := len(args) == 1
	hasFrom := fromFlag != ""
	hasTo := toFlag != ""

	if hasID && (hasFrom || hasTo) {
		return apperrors.NewInvalidInputError(
			"payout ID and --from/--to flags are mutually exclusive",
		)
	}

	if hasFrom != hasTo {
		return apperrors.NewInvalidInputError(
			"both --from and --to must be provided for period reconciliation",
		)
	}

	if !hasID && !hasFrom {
		return apperrors.NewInvalidInputError(
			"provide a payout ID or --from/--to date range",
		)
	}

	if hasID {
		return validatePayoutID(args[0])
	}

	from, err := parseDateFlag(fromFlag)
	if err != nil {
		return apperrors.NewInvalidInputError(
			fmt.Sprintf("invalid --from date: %s", err),
		)
	}

	to, err := parseDateFlag(toFlag)
	if err != nil {
		return apperrors.NewInvalidInputError(
			fmt.Sprintf("invalid --to date: %s", err),
		)
	}

	if to.Before(from) {
		return apperrors.NewInvalidInputError(
			fmt.Sprintf("--to (%s) must not be before --from (%s)", toFlag, fromFlag),
		)
	}

	return nil
}

const dateFormat = "2006-01-02"

// parseDateFlag parses a YYYY-MM-DD string as a UTC date.
func parseDateFlag(value string) (time.Time, error) {
	return time.Parse(dateFormat, value)
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
