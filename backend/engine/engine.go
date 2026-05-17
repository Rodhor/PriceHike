package engine

import (
	"PriceHike/backend/domain"
	"fmt"
	"time"
)

func ApplyRules(rules []domain.Rule, ds domain.Dataset, decimalPlaces int) (domain.Dataset, []domain.RuleStat, error) {
	out := ds.Clone()
	stats := make([]domain.RuleStat, 0, len(rules))
	for i, rule := range rules {
		if err := rule.Validate(out.ColumnCount()); err != nil {
			return domain.Dataset{}, stats, fmt.Errorf("rule #%d: %w", i+1, err)
		}
		cols := targetColumns(rule, out.ColumnCount())
		stat := domain.RuleStat{RuleID: rule.ID}
		for itemIdx := range out.Items {
			if !Matches(out.Items[itemIdx], rule.Filter) {
				continue
			}
			stat.ItemsMatched++
			for _, c := range cols {
				before := out.Items[itemIdx].Prices[c]
				p := ApplyAction(before, rule.Action)
				p = domain.ApplyDecimalSettings(p, rule.DecimalSettings)
				if !before.Equal(p) {
					stat.CellsChanged++
				}
				out.Items[itemIdx].Prices[c] = p
			}
		}
		if stat.ItemsMatched == 0 {
			stat.Warning = "matched 0 items"
		}
		stats = append(stats, stat)

	}
	out.ExportPrecision = rulesPlaces(rules, out.ColumnCount(), decimalPlaces)

	return out, stats, nil

}

func Preview(rules []domain.Rule, ds domain.Dataset, sampleSize int, decimalPlaces int) (domain.PreviewResult, error) {
	if sampleSize <= 0 {
		sampleSize = 500
	}
	modified, stats, err := ApplyRules(rules, ds, decimalPlaces)
	if err != nil {
		return domain.PreviewResult{}, err
	}
	places := modified.ExportPrecision
	affected := make([]domain.PreviewRow, 0, sampleSize)
	unaffected := make([]domain.PreviewRow, 0, sampleSize)
	totalAffectedItems := 0
	totalAffectedCells := 0
	for i := range ds.Items {
		cells := changedCells(ds.Items[i], modified.Items[i], places)
		rowChanged := false
		for _, c := range cells {
			if c.Changed {
				totalAffectedCells++
				rowChanged = true
			}
		}

		row := domain.PreviewRow{
			ItemNumber: ds.Items[i].ItemNumber,
			Cells:      cells,
		}
		if rowChanged {
			totalAffectedItems++
			if len(affected) < sampleSize {
				affected = append(affected, row)
			}
		} else if len(unaffected) < sampleSize {
			unaffected = append(unaffected, row)
		}
	}

	rows := affected
	if remaining := sampleSize - len(affected); remaining > 0 {
		if remaining > len(unaffected) {
			remaining = len(unaffected)
		}
		rows = append(rows, unaffected[:remaining]...)
	}

	return domain.PreviewResult{
		Rows:               rows,
		TotalAffectedItems: totalAffectedItems,
		TotalAffectedCells: totalAffectedCells,
		TotalItems:         len(ds.Items),
		SampleSize:         sampleSize,
		PerRuleStats:       stats,
		GeneratedAt:        time.Now().UTC().Format(time.RFC3339),
	}, nil
}
