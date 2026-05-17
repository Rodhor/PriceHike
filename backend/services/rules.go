package services

import (
	"PriceHike/backend/domain"
	rulesio "PriceHike/backend/io/rules"
	"fmt"
)

type RulesService struct{}

func NewRulesService() *RulesService {
	return &RulesService{}
}

func (rs *RulesService) ImportRules(path string) ([]domain.Rule, error) {
	rules, err := rulesio.Read(path)
	if err != nil {
		return nil, err
	}
	for i, r := range rules {
		if err := r.ValidateShape(); err != nil {
			return nil, fmt.Errorf("rule #%d (%s): %w", i+1, r.ID, err)
		}
	}
	return rules, nil
}

func (rs *RulesService) ExportRules(rules []domain.Rule, destPath string) error {
	return rulesio.Write(destPath, rules)
}
