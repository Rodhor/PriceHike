package engine

import (
	"PriceHike/backend/domain"
	"path/filepath"

	"github.com/shopspring/decimal"
)

func Matches(item domain.Item, filter domain.RuleFilter) bool {
	if filter.ApplyToAll {
		return true
	}

	hasWildcard := filter.Wildcard != ""
	hasLessThan := filter.LessThan != nil
	hasGreaterThan := filter.GreaterThan != nil

	if !hasWildcard && !hasLessThan && !hasGreaterThan {
		return false
	}

	if hasWildcard && !matchWildcard(item.ItemNumber, filter.Wildcard) {
		return false
	}
	if hasLessThan && !matchLessThan(item.Prices, filter.LessThan) {
		return false
	}
	if hasGreaterThan && !matchGreaterThan(item.Prices, filter.GreaterThan) {
		return false
	}

	return true
}

func matchWildcard(itemNumber, pattern string) bool {
	matched, err := filepath.Match(pattern, itemNumber)
	if err != nil {
		return false
	}
	return matched

}

func matchLessThan(prices []decimal.Decimal, lt *domain.LessThanFilter) bool {
	if lt.Column < 0 || lt.Column >= len(prices) {
		return false
	}
	return prices[lt.Column].LessThan(lt.Value)
}

func matchGreaterThan(prices []decimal.Decimal, gt *domain.GreaterThanFilter) bool {
	if gt.Column < 0 || gt.Column >= len(prices) {
		return false
	}
	return prices[gt.Column].GreaterThan(gt.Value)
}
