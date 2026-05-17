package csvio

import (
	"PriceHike/backend/domain"
	"encoding/csv"
	"fmt"
	"os"
)

func Write(path string, ds domain.Dataset, comma rune) error {

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if comma != 0 {
		w.Comma = comma
	}
	defer w.Flush()

	header := append([]string{ds.ItemHeader}, ds.PriceHeaders...)
	if err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for i, item := range ds.Items {
		row := make([]string, 0, 1+len(item.Prices))
		row = append(row, item.ItemNumber)
		for c, p := range item.Prices {
			row = append(row, domain.FormatPrice(p, ds.ExportPrecision, c))
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write row %d: %w", i, err)
		}
	}
	return w.Error()
}
