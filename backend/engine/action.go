package engine

import (
	"PriceHike/backend/domain"

	"github.com/shopspring/decimal"
)

var hundred = decimal.NewFromInt(100)

func ApplyAction(price decimal.Decimal, action domain.RuleAction) decimal.Decimal {

	switch action.Type {
	case domain.ActionAddPercent:

		return applyPercent(price, action.Value, true)
	case domain.ActionSubtractPercent:
		return applyPercent(price, action.Value, false)
	case domain.ActionAddFixed:
		return price.Add(action.Value)
	case domain.ActionSubtractFixed:
		return price.Sub(action.Value)
	case domain.ActionRoundUp:
		return price.Ceil()
	case domain.ActionRoundDown:
		return price.Floor()
	default:
		return price
	}
}

func applyPercent(price, pct decimal.Decimal, add bool) decimal.Decimal {
	delta := price.Mul(pct).Div(hundred)
	if add {
		return price.Add(delta)
	}
	return price.Sub(delta)
}
