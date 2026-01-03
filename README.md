# stripe-payout-reconciler

[![CI](https://github.com/navfa/stripe-payout-reconciler/actions/workflows/ci.yml/badge.svg)](https://github.com/navfa/stripe-payout-reconciler/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/navfa/stripe-payout-reconciler)](https://goreportcard.com/report/github.com/navfa/stripe-payout-reconciler)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](go.mod)
[![Release](https://img.shields.io/badge/Release-v1.0.0-blue)](https://github.com/navfa/stripe-payout-reconciler/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A command-line tool that fetches Stripe payouts and their balance transactions, outputting them in CSV, JSON, or JSONL for reconciliation with your accounting records.

## Installation

### From source

```sh
go install github.com/navfa/stripe-payout-reconciler/cmd/stripe-payout-reconciler@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/navfa/stripe-payout-reconciler/releases). Binaries are available for Linux and macOS (amd64 and arm64).

## Quick Start

```sh
# Set your Stripe API key
export STRIPE_API_KEY=sk_test_...

# Reconcile a single payout
stripe-payout-reconciler payout po_1ABC2DEF3GHI

# Export a month of payouts as JSON
stripe-payout-reconciler payout --from 2024-01-01 --to 2024-01-31 --format json

# Pipe JSONL to jq for filtering
stripe-payout-reconciler payout po_1ABC2DEF3GHI --format jsonl | jq 'select(.type == "charge")'
```

## Usage

### Single payout

```sh
stripe-payout-reconciler payout <payout-id> [flags]
```

Fetches all balance transactions for a specific payout and outputs them.

### Period reconciliation

```sh
stripe-payout-reconciler payout --from <YYYY-MM-DD> --to <YYYY-MM-DD> [flags]
```

Fetches all payouts in the date range and outputs their combined transactions. Dates are interpreted as UTC.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` | | Stripe API key (overrides `STRIPE_API_KEY` env var) |
| `--format` | `csv` | Output format: `csv`, `json`, or `jsonl` |
| `--from` | | Start date, inclusive (UTC, `YYYY-MM-DD`) |
| `--to` | | End date, inclusive (UTC, `YYYY-MM-DD`) |

### Configuration

The Stripe API key can be provided in two ways (in order of precedence):

1. `--api-key` flag
2. `STRIPE_API_KEY` environment variable

All data goes to stdout, all diagnostics to stderr. Pipe output to a file or another tool:

```sh
stripe-payout-reconciler payout po_xxx --format csv > payout.csv
```

## Output Formats

### CSV

```csv
payout_id,transaction_id,type,amount,fee,net,currency,created,description
po_abc123,txn_def456,charge,150.00,4.35,145.65,usd,2024-01-15T14:30:00Z,Payment for invoice #1234
```

### JSON

```json
[
  {
    "payout_id": "po_abc123",
    "transaction_id": "txn_def456",
    "type": "charge",
    "amount": "150.00",
    "fee": "4.35",
    "net": "145.65",
    "currency": "usd",
    "created": "2024-01-15T14:30:00Z",
    "description": "Payment for invoice #1234"
  }
]
```

### JSONL

```jsonl
{"payout_id":"po_abc123","transaction_id":"txn_def456","type":"charge","amount":"150.00","fee":"4.35","net":"145.65","currency":"usd","created":"2024-01-15T14:30:00Z","description":"Payment for invoice #1234"}
```

Amounts are formatted as decimal strings (not floats) to avoid rounding errors in financial data. Zero-decimal currencies like JPY are output as integers.

## Architecture

```
cmd/stripe-payout-reconciler/   CLI entry point (Cobra commands)
    |
    +-- internal/config/        API key resolution (flag > env var)
    +-- internal/stripe/        Stripe API client (interface + implementation)
    +-- internal/format/        Output formatters (CSV, JSON, JSONL)
    +-- internal/model/         Domain types (Payout, Record)
    +-- internal/errors/        Structured error types with exit codes
```

The `model` package imports nothing from the rest of the project. The `stripe` package encapsulates all stripe-go dependencies. Formatters depend only on `model`.

## Development

```sh
make build                                          # compile the binary
make run ARGS="payout po_xxx --format json"         # build and run with arguments
make test                                           # run tests with race detector
make check                                          # fmt + tidy + verify + vet + lint + test
make coverage                                       # test with coverage report
make install                                        # install to $GOBIN
make clean                                          # remove build artifacts
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development setup and guidelines.

## License

[MIT](LICENSE)
