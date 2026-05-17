package engine

import "PriceHike/backend/domain"

func targetColumns(rule domain.Rule, columnCount int) []int {
	if len(rule.Columns) == 0 {
		out := make([]int, columnCount)
		for i := range out {
			out[i] = i
		}
		return out
	}

	out := make([]int, 0, len(rule.Columns))
	for _, c := range rule.Columns {
		if c >= 0 && c < columnCount {
			out = append(out, c)
		}
	}

	return out
}
