// Package config resolves runtime configuration from flags and environment
// variables. The precedence order is: flag value > environment variable > error.
//
// There is no config file — flags and env vars cover the use cases for a
// CLI tool that runs in scripts, CI pipelines, and interactive terminals.
// See ADR-002 for the reasoning.
package config

import (
	"os"

	apperrors "github.com/navfa/stripe-payout-reconciler/internal/errors"
)

const envKeyAPIKey = "STRIPE_API_KEY"

// Config holds the resolved runtime configuration.
type Config struct {
	APIKey string
}

// Load resolves the API key with precedence: flag > env var > error.
func Load(flagAPIKey string) (Config, error) {
	apiKey := flagAPIKey
	if apiKey == "" {
		apiKey = os.Getenv(envKeyAPIKey)
	}

	if apiKey == "" {
		return Config{}, apperrors.NewInvalidInputError(
			"API key is required: use --api-key flag or set STRIPE_API_KEY environment variable",
		)
	}

	return Config{APIKey: apiKey}, nil
}
