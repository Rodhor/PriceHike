package domain

import (
	"github.com/shopspring/decimal"
)

func ApplyDecimalSettings(price decimal.Decimal, settings DecimalSettings) decimal.Decimal {
	switch settings.Mode {
	case ModeFixed:
		if settings.Places < 0 {
			return price
		}
		return price.Round(int32(settings.Places))
	case ModeKeep:
		return price

	// If nothing is matched, return the price as-is
	default:
		return price
	}
}

func FormatPrice(price decimal.Decimal, precision []int, col int) string {
	places := 2
	if col < len(precision) {
		places = precision[col]
	}
	return price.StringFixed(int32(places))
}

func FormatDiff(oldV, newV decimal.Decimal, places int) string {
	if places < 0 {
		places = 0
	}
	diff := newV.Sub(oldV)
	formatted := diff.StringFixed(int32(places))

	if diff.IsPositive() {
		return "+" + formatted
	}
	return formatted

}
