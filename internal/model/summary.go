package model

// TypeSummary holds aggregated totals for one RecordType.
type TypeSummary struct {
	Type   RecordType
	Count  int
	Amount int64
	Fee    int64
	Net    int64
}

// Summary holds aggregated totals for a set of records in one currency.
type Summary struct {
	Currency string
	ByType   []TypeSummary
	Total    TypeSummary
}

// canonicalOrder defines the display order for record types.
var canonicalOrder = []RecordType{
	RecordTypeCharge,
	RecordTypeRefund,
	RecordTypeFee,
	RecordTypeDispute,
	RecordTypeAdjustment,
	RecordTypeOther,
}

// Summarize aggregates records by currency and type. Returns one Summary
// per distinct currency found. Within each Summary, ByType entries follow
// canonical order and include all types (zero-valued if absent).
func Summarize(records []Record) []Summary {
	if len(records) == 0 {
		return nil
	}

	// Group records by currency.
	byCurrency := make(map[string][]Record)
	var currencyOrder []string
	for _, r := range records {
		if _, seen := byCurrency[r.Currency]; !seen {
			currencyOrder = append(currencyOrder, r.Currency)
		}
		byCurrency[r.Currency] = append(byCurrency[r.Currency], r)
	}

	summaries := make([]Summary, 0, len(currencyOrder))
	for _, cur := range currencyOrder {
		summaries = append(summaries, summarizeCurrency(cur, byCurrency[cur]))
	}
	return summaries
}

func summarizeCurrency(currency string, records []Record) Summary {
	byType := make(map[RecordType]*TypeSummary)
	for _, rt := range canonicalOrder {
		byType[rt] = &TypeSummary{Type: rt}
	}

	var total TypeSummary
	for _, r := range records {
		ts, ok := byType[r.Type]
		if !ok {
			ts = byType[RecordTypeOther]
		}
		ts.Count++
		ts.Amount += r.Amount
		ts.Fee += r.Fee
		ts.Net += r.Net

		total.Count++
		total.Amount += r.Amount
		total.Fee += r.Fee
		total.Net += r.Net
	}

	result := make([]TypeSummary, len(canonicalOrder))
	for i, rt := range canonicalOrder {
		result[i] = *byType[rt]
	}

	return Summary{
		Currency: currency,
		ByType:   result,
		Total:    total,
	}
}
