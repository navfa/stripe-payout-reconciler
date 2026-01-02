// Package main is the entry point for the stripe-payout-reconciler CLI.
package main

import (
	"fmt"
	"os"

	apperrors "github.com/paco/stripe-payout-reconciler/internal/errors"
)

// version is set at build time via ldflags. Development builds show "dev".
var version = "dev"

func main() {
	if err := run(); err != nil {
		printError(err)
		os.Exit(apperrors.ExitCode(err))
	}
}

// run builds the command tree and executes it.
func run() error {
	rootCmd := newRootCmd()
	return rootCmd.Execute()
}

// printError writes a user-facing error message to stderr, preferring
// UserMessage() when available.
func printError(err error) {
	type userMessager interface {
		UserMessage() string
	}

	if messager, ok := err.(userMessager); ok {
		fmt.Fprintln(os.Stderr, "Error:", messager.UserMessage())
		return
	}

	fmt.Fprintln(os.Stderr, "Error:", err.Error())
}
