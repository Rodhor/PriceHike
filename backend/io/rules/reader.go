package rulesio

import (
	"PriceHike/backend/domain"
	"encoding/json"
	"fmt"
	"os"
)

type fileShape struct {
	Rules []domain.Rule `json:"rules"`
}

func Read(path string) ([]domain.Rule, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var f fileShape
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	return f.Rules, nil

}
