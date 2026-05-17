package excelio

import (
	"PriceHike/backend/domain"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func Read(path string) (*domain.Dataset, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	if f.SheetCount == 0 {
		return nil, fmt.Errorf("open %s: no sheets", path)
	}
	sheet := f.GetSheetName(0)

	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows for %s: %w", sheet, err)
	}

	if !rows.Next() {
		return nil, fmt.Errorf("read header: sheet %s is empty", sheet)
	}
	header, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read header columns: %w", err)
	}
	if len(header) == 0 {
		return nil, fmt.Errorf("read header: first row is empty")
	}

	var items []domain.Item
	for rows.Next() {
		rec, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", len(items)+1, err)
		}

		if len(rec) == 0 {
			continue
		}

		var item domain.Item
		for i, v := range rec {
			if i == 0 {
				item.ItemNumber = v
			} else {
				if v == "" {
					item.Prices = append(item.Prices, decimal.Zero)
					continue
				}

				price, err := decimal.NewFromString(v)
				if err != nil {
					return nil, fmt.Errorf("parse price %s at row %d: %w", v, len(items)+1, err)
				}
				item.Prices = append(item.Prices, price)
			}
		}
		items = append(items, item)
	}

	if err := rows.Error(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return &domain.Dataset{
		ItemHeader:   header[0],
		PriceHeaders: header[1:],
		Items:        items,
	}, nil

}
