package format

import (
	"fmt"
	"strconv"
	"strings"
)

// zeroDecimalCurrencies lists ISO 4217 currencies that have no decimal places.
// Stripe treats these as whole units (e.g., 1500 JPY is ¥1500, not ¥15.00).
var zeroDecimalCurrencies = map[string]bool{
	"bif": true,
	"clp": true,
	"djf": true,
	"gnf": true,
	"jpy": true,
	"kmf": true,
	"krw": true,
	"mga": true,
	"pyg": true,
	"rwf": true,
	"ugx": true,
	"vnd": true,
	"vuv": true,
	"xaf": true,
	"xof": true,
	"xpf": true,
}

// FormatAmount converts an integer amount in the smallest currency unit to a
// decimal string appropriate for the given currency. Zero-decimal currencies
// (e.g., JPY) are returned as plain integers. Standard currencies default to
// 2 decimal places. Unknown currencies default to 2 decimals per ADR-006.
func FormatAmount(amount int64, currency string) string {
	code := strings.ToLower(currency)

	if zeroDecimalCurrencies[code] {
		return strconv.FormatInt(amount, 10)
	}

	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	return fmt.Sprintf("%s%d.%02d", sign, amount/100, amount%100)
}
