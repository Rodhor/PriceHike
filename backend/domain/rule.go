package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Rule struct {
	ID              string          `json:"id"`
	Filter          RuleFilter      `json:"filter"`
	Columns         []int           `json:"columns"`
	Action          RuleAction      `json:"action"`
	DecimalSettings DecimalSettings `json:"decimalSettings"`
}

type RuleFilter struct {
	ApplyToAll  bool               `json:"applyToAll"`
	Wildcard    string             `json:"wildcard,omitempty"`
	LessThan    *LessThanFilter    `json:"lessThan,omitempty"`
	GreaterThan *GreaterThanFilter `json:"greaterThan,omitempty"`
}

type RuleAction struct {
	Type  RuleActionType  `json:"type"`
	Value decimal.Decimal `json:"value" ts_type:"string"`
}

type RuleActionType string

const (
	ActionAddPercent      RuleActionType = "addPercent"
	ActionSubtractPercent RuleActionType = "subtractPercent"
	ActionAddFixed        RuleActionType = "addFixed"
	ActionSubtractFixed   RuleActionType = "subtractFixed"
	ActionRoundUp         RuleActionType = "roundUp"
	ActionRoundDown       RuleActionType = "roundDown"
)

type DecimalSettingsMode string

const (
	ModeKeep  DecimalSettingsMode = "keep"
	ModeFixed DecimalSettingsMode = "fixed"
)

type DecimalSettings struct {
	Mode   DecimalSettingsMode `json:"mode"`
	Places int                 `json:"places"`
}

type LessThanFilter struct {
	Column int             `json:"column"`
	Value  decimal.Decimal `json:"value" ts_type:"string"`
}

type GreaterThanFilter struct {
	Column int             `json:"column"`
	Value  decimal.Decimal `json:"value" ts_type:"string"`
}

func (r *Rule) ValidateShape() error {
	if r.ID == "" {
		return fmt.Errorf("%w: ID must be non-empty", ErrInvalidRule)
	}
	if !r.Filter.ApplyToAll && (r.Filter.GreaterThan == nil && r.Filter.LessThan == nil && r.Filter.Wildcard == "") {
		return fmt.Errorf("%w: filter must be specified", ErrInvalidRule)
	}
	if r.Filter.ApplyToAll && (r.Filter.GreaterThan != nil || r.Filter.LessThan != nil || r.Filter.Wildcard != "") {
		return fmt.Errorf("%w: filter must not be specified when applyToAll is true", ErrInvalidRule)
	}
	switch r.Action.Type {
	case ActionAddFixed, ActionAddPercent, ActionRoundDown, ActionRoundUp, ActionSubtractFixed:
		break
	case ActionSubtractPercent:
		hundred := decimal.NewFromInt(100)
		zero := decimal.Zero
		if r.Action.Value.GreaterThan(hundred) || r.Action.Value.LessThanOrEqual(zero) {
			return fmt.Errorf("%w: value must be between 0 and 100", ErrInvalidRule)
		}
	default:
		return fmt.Errorf("%w: invalid action type", ErrInvalidRule)
	}

	if r.Action.RequiresValue() && r.Action.Value.IsZero() {
		return fmt.Errorf("%w: value must be non-zero", ErrInvalidRule)
	}
	switch r.DecimalSettings.Mode {
	case ModeKeep:
	case ModeFixed:
		if r.DecimalSettings.Places < 0 || r.DecimalSettings.Places > 10 {
			return fmt.Errorf("%w: decimal places must be between 0 and 10", ErrInvalidRule)
		}
	default:
		return fmt.Errorf("%w: invalid decimal mode", ErrInvalidRule)
	}

	return nil
}

func (r *Rule) Validate(columnCount int) error {
	if err := r.ValidateShape(); err != nil {
		return err
	}
	for _, colIdx := range r.Columns {
		if colIdx < 0 || colIdx >= columnCount {
			return fmt.Errorf("%w: column index must be within range", ErrInvalidRule)
		}
	}
	if lt := r.Filter.LessThan; lt != nil {
		if lt.Column < 0 || lt.Column >= columnCount {
			return fmt.Errorf("%w: LessThanFilter column index %d out of range (have %d columns)", ErrInvalidRule, lt.Column, columnCount)
		}
	}
	if gt := r.Filter.GreaterThan; gt != nil {
		if gt.Column < 0 || gt.Column >= columnCount {
			return fmt.Errorf("%w: GreaterThanFilter column index %d out of range (have %d columns)", ErrInvalidRule, gt.Column, columnCount)
		}
	}
	return nil
}

func (a RuleAction) IsRounding() bool {
	return a.Type == ActionRoundUp || a.Type == ActionRoundDown
}

func (a RuleAction) RequiresValue() bool {
	return !a.IsRounding()
}
