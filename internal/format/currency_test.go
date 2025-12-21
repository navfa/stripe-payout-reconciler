package format

import "testing"

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		want     string
	}{
		{name: "USD positive", amount: 15000, currency: "usd", want: "150.00"},
		{name: "USD negative", amount: -15000, currency: "usd", want: "-150.00"},
		{name: "USD zero", amount: 0, currency: "usd", want: "0.00"},
		{name: "USD small amount", amount: 5, currency: "usd", want: "0.05"},
		{name: "USD one cent", amount: 1, currency: "usd", want: "0.01"},
		{name: "JPY positive", amount: 1500, currency: "jpy", want: "1500"},
		{name: "JPY negative", amount: -1500, currency: "jpy", want: "-1500"},
		{name: "JPY zero", amount: 0, currency: "jpy", want: "0"},
		{name: "KRW zero-decimal", amount: 50000, currency: "krw", want: "50000"},
		{name: "unknown currency defaults to 2 decimals", amount: 9999, currency: "xyz", want: "99.99"},
		{name: "uppercase USD", amount: 15000, currency: "USD", want: "150.00"},
		{name: "mixed case Jpy", amount: 1500, currency: "Jpy", want: "1500"},
		{name: "EUR positive", amount: 1234, currency: "eur", want: "12.34"},
		{name: "GBP negative", amount: -50, currency: "gbp", want: "-0.50"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := FormatAmount(testCase.amount, testCase.currency)
			if got != testCase.want {
				t.Errorf("FormatAmount(%d, %q) = %q, want %q",
					testCase.amount, testCase.currency, got, testCase.want)
			}
		})
	}
}
