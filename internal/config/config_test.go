package config_test

import (
	"errors"
	"testing"

	"github.com/paco/stripe-payout-reconciler/internal/config"
	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		flagValue  string
		envValue   string
		wantAPIKey string
		wantErr    bool
	}{
		{
			name:       "flag value takes precedence over env",
			flagValue:  "sk_test_from_flag",
			envValue:   "sk_test_from_env",
			wantAPIKey: "sk_test_from_flag",
		},
		{
			name:       "falls back to env when flag is empty",
			flagValue:  "",
			envValue:   "sk_test_from_env",
			wantAPIKey: "sk_test_from_env",
		},
		{
			name:      "returns error when both are empty",
			flagValue: "",
			envValue:  "",
			wantErr:   true,
		},
		{
			name:       "flag value used when env is not set",
			flagValue:  "sk_test_flag_only",
			envValue:   "",
			wantAPIKey: "sk_test_flag_only",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("STRIPE_API_KEY", testCase.envValue)

			cfg, err := config.Load(testCase.flagValue)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("Load() returned nil error, want error")
				}
				var inputErr *apperrors.InvalidInputError
				if !errors.As(err, &inputErr) {
					t.Errorf("Load() error type = %T, want *errors.InvalidInputError", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}
			if cfg.APIKey != testCase.wantAPIKey {
				t.Errorf("Load().APIKey = %q, want %q", cfg.APIKey, testCase.wantAPIKey)
			}
		})
	}
}
