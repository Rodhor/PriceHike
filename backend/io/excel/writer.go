package excelio

import (
	"PriceHike/backend/domain"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func Write(path string, ds domain.Dataset) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	const sheet = "prices"
	f.SetSheetName(f.GetSheetName(0), sheet)

	header := append([]string{ds.ItemHeader}, ds.PriceHeaders...)
	for c, h := range header {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		if err := f.SetCellStr(sheet, cell, h); err != nil {
			return fmt.Errorf("write header col %d: %w", c, err)
		}
	}

	for i, item := range ds.Items {
		rowIdx := i + 2
		key, _ := excelize.CoordinatesToCellName(1, rowIdx)
		if err := f.SetCellStr(sheet, key, item.ItemNumber); err != nil {
			return fmt.Errorf("write row %d key: %w", rowIdx, err)
		}
		for c := range ds.PriceHeaders {
			cell, _ := excelize.CoordinatesToCellName(c+2, rowIdx)
			var price decimal.Decimal
			if c < len(item.Prices) {
				price = item.Prices[c]
			}
			if err := f.SetCellStr(sheet, cell, domain.FormatPrice(price, ds.ExportPrecision, c)); err != nil {
				return fmt.Errorf("write row %d col %d: %w", rowIdx, c+2, err)
			}
		}
	}
	return f.SaveAs(path)
}
