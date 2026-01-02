# Example Output

These files show the output formats produced by `stripe-payout-reconciler`. All data is fictitious.

- `single-payout.csv` — A single payout with its balance transactions in CSV format
- `single-payout.json` — The same payout in JSON format
- `period-reconciliation.jsonl` — Two payouts from a date range query in JSONL format

## Reproducing

```sh
# Single payout
stripe-payout-reconciler payout po_xxx --format csv > single-payout.csv
stripe-payout-reconciler payout po_xxx --format json > single-payout.json

# Period reconciliation
stripe-payout-reconciler payout --from 2024-01-14 --to 2024-01-17 --format jsonl > period-reconciliation.jsonl
```
