# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-01-05

### Added

- `--summary` flag: prints aggregated breakdown by transaction type (charges, refunds, fees, disputes, adjustments) to stderr
- Multi-currency support in summary: groups totals by currency when payouts span multiple currencies

## [1.0.0] - 2026-01-02

### Added

- Fetch balance transactions for a single payout (`payout <id>`)
- Period reconciliation with `--from` and `--to` date range flags
- Output formats: CSV, JSON, and JSONL (`--format` flag)
- Decimal string formatting for financial amounts (no floating-point rounding)
- Zero-decimal currency support (JPY, KRW, etc.)
- Stripe API key via `--api-key` flag or `STRIPE_API_KEY` environment variable
- Automatic retry with backoff for rate-limited requests (via stripe-go SDK)
- Clean cancellation on Ctrl+C
- Pre-built binaries for Linux and macOS (amd64, arm64) via GoReleaser
