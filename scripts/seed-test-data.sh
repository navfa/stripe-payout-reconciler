#!/usr/bin/env bash
# Creates test transactions on a Stripe test account for local testing.
# Usage: STRIPE_API_KEY=sk_test_... ./scripts/seed-test-data.sh
set -euo pipefail

: "${STRIPE_API_KEY:?Set STRIPE_API_KEY to your sk_test_... key}"

API="https://api.stripe.com/v1"

stripe_post() {
  local response
  response=$(curl -s -X POST "$API/$1" \
    -u "$STRIPE_API_KEY:" \
    -d "$2")
  local id
  id=$(echo "$response" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))" 2>/dev/null)
  if [ -z "$id" ]; then
    echo "Error calling $1:" >&2
    echo "$response" >&2
    exit 1
  fi
  echo "$id"
}

charge() {
  local amount=$1
  stripe_post charges "amount=${amount}&currency=usd&source=tok_visa&description=Test+charge+${amount}"
}

echo "Creating test charges..."

CH1=$(charge 15000)
echo "  Charge 1: \$150.00 ($CH1)"

CH2=$(charge 8999)
echo "  Charge 2: \$89.99 ($CH2)"

CH3=$(charge 25000)
echo "  Charge 3: \$250.00 ($CH3)"

CH4=$(charge 1250)
echo "  Charge 4: \$12.50 ($CH4)"

echo ""
echo "Creating a refund on charge 2..."
REF=$(stripe_post refunds "charge=$CH2")
echo "  Refund: $REF (full refund of \$89.99)"

echo ""
echo "Done! Created 4 charges and 1 refund."
echo ""
echo "Stripe will bundle these into a payout automatically (usually within 2 days in test mode)."
echo "Check your payouts:"
echo "  make run ARGS=\"payout --from $(date -v-7d +%Y-%m-%d 2>/dev/null || date -d '7 days ago' +%Y-%m-%d) --to $(date +%Y-%m-%d) --summary\""
