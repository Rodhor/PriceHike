package csvio

import (
	"PriceHike/backend/domain"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/shopspring/decimal"
)

func Read(path string, comma rune) (*domain.Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	if comma == 0 {
		reader.Comma = ','
	} else {
		reader.Comma = comma
	}
	reader.ReuseRecord = false
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	var items []domain.Item
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", len(items)+1, err)
		}
		var item domain.Item
		for i, v := range rec {
			if i == 0 {
				item.ItemNumber = v
			} else {
				price, err := decimal.NewFromString(v)
				if err != nil {
					return nil, fmt.Errorf("parse price %s: %w", v, err)
				}
				item.Prices = append(item.Prices, price)
			}
		}
		items = append(items, item)
	}

	return &domain.Dataset{
		ItemHeader:   header[0],
		PriceHeaders: header[1:],
		Items:        items,
	}, nil
}
