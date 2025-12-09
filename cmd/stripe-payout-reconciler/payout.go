package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/paco/stripe-payout-reconciler/internal/config"
	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
)

const payoutIDPrefix = "po_"

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

	// TODO: fetch payout data from Stripe and format output.
	keyPreview := redactAPIKey(cfg.APIKey)
	fmt.Fprintf(
		os.Stderr,
		"Validated payout %s with API key %s\n",
		payoutID, keyPreview,
	)

	return nil
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
