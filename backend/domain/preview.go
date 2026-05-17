package domain

import "github.com/shopspring/decimal"

type PreviewResult struct {
	TotalAffectedItems int          `json:"totalAffectedItems"`
	TotalAffectedCells int          `json:"totalAffectedCells"`
	TotalItems         int          `json:"totalItems"`
	SampleSize         int          `json:"sampleSize"`
	Rows               []PreviewRow `json:"rows"`
	PerRuleStats       []RuleStat   `json:"perRuleStats"`
	GeneratedAt        string       `json:"generatedAt"`
}

type RuleStat struct {
	RuleID       string `json:"ruleId"`
	ItemsMatched int    `json:"itemsMatched"`
	CellsChanged int    `json:"cellsChanged"`
	Warning      string `json:"warning,omitempty"`
}

type PreviewRow struct {
	ItemNumber string        `json:"itemNumber"`
	Cells      []PreviewCell `json:"cells"`
}

type PreviewCell struct {
	Column   int    `json:"column"`
	Original string `json:"original"`
	NewValue string `json:"newValue"`
	Diff     string `json:"diff"`
	Changed  bool   `json:"changed"`
}

func NewPreviewCell(col int, oldV, newV decimal.Decimal, places int) PreviewCell {
	changed := !oldV.Equal(newV)

	return PreviewCell{
		Column:   col,
		Original: oldV.StringFixed(int32(places)),
		NewValue: newV.StringFixed(int32(places)),
		Diff:     FormatDiff(oldV, newV, places),
		Changed:  changed,
	}
}
