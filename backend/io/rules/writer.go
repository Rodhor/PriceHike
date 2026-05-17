package rulesio

import (
	"PriceHike/backend/domain"
	"encoding/json"
	"fmt"
	"os"
)

func Write(path string, rules []domain.Rule) error {
	body, err := json.MarshalIndent(fileShape{Rules: rules}, "", " ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
