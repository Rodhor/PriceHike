package engine

import "PriceHike/backend/domain"

func rulesPlaces(rules []domain.Rule, columnCount int, defaultPlaces int) []int {
	places := make([]int, columnCount)
	for i := range places {
		places[i] = defaultPlaces
	}
	for _, r := range rules {
		if r.DecimalSettings.Mode != "fixed" {
			continue
		}
		cols := targetColumns(r, columnCount)
		for _, c := range cols {
			places[c] = r.DecimalSettings.Places
		}
	}
	return places
}

func changedCells(orig, modified domain.Item, places []int) []domain.PreviewCell {
	cells := make([]domain.PreviewCell, 0, len(orig.Prices))
	for c := range orig.Prices {
		cells = append(cells, domain.NewPreviewCell(c, orig.Prices[c], modified.Prices[c], places[c]))
	}
	return cells
}
