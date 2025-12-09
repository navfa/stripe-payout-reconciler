package main

import (
	"github.com/spf13/cobra"
)

var apiKeyFlag string

// newRootCmd creates the top-level CLI command with the --api-key
// persistent flag and all subcommands.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "stripe-payout-reconciler",
		Short: "Reconcile Stripe payouts with their balance transactions",
		Long: `stripe-payout-reconciler fetches payout details from the Stripe API
and outputs the associated balance transactions in a structured format
(CSV, JSON, or JSONL) for reconciliation with your accounting records.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(
		&apiKeyFlag,
		"api-key",
		"",
		"Stripe API key (overrides STRIPE_API_KEY env var)",
	)

	rootCmd.AddCommand(newPayoutCmd())

	return rootCmd
}
