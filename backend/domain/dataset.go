package domain

import "github.com/shopspring/decimal"

type Dataset struct {
	ItemHeader      string   `json:"itemHeader"`
	PriceHeaders    []string `json:"priceHeaders"`
	Items           []Item   `json:"items"`
	ExportPrecision []int    `json:"exportPrecision"`
}

type Item struct {
	ItemNumber string            `json:"itemNumber"`
	Prices     []decimal.Decimal `json:"prices" ts_type:"string[]"`
}

func (d *Dataset) ColumnCount() int {
	return len(d.PriceHeaders)
}

func (d *Dataset) RowCount() int {
	return len(d.Items)
}

func (d *Dataset) ColumnIndex(header string) (int, bool) {
	for i, h := range d.PriceHeaders {
		if h == header {
			return i, true
		}
	}
	return -1, false
}

func (d *Dataset) Clone() Dataset {
	if d == nil {
		return Dataset{}
	}
	clone := *d

	if d.PriceHeaders != nil {
		clone.PriceHeaders = make([]string, len(d.PriceHeaders))
		copy(clone.PriceHeaders, d.PriceHeaders)
	}

	if d.Items != nil {
		clone.Items = make([]Item, len(d.Items))
		for i, originalItem := range d.Items {
			clonedItem := originalItem

			if originalItem.Prices != nil {
				clonedItem.Prices = make([]decimal.Decimal, len(originalItem.Prices))
				copy(clonedItem.Prices, originalItem.Prices)
			}
			clone.Items[i] = clonedItem
		}
	}
	if d.ExportPrecision != nil {
		clone.ExportPrecision = make([]int, len(d.ExportPrecision))
		copy(clone.ExportPrecision, d.ExportPrecision)
	}
	return clone
}
