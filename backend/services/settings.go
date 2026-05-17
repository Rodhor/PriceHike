package services

import (
	"PriceHike/backend/domain"
	"sync"
)

type SettingsService struct {
	mu       sync.RWMutex
	settings domain.Settings
}

func NewSettingsService() *SettingsService {
	return &SettingsService{
		settings: domain.DefaultSettings(),
	}
}

func (s *SettingsService) GetSettings() domain.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *SettingsService) SetSettings(newSettings domain.Settings) error {
	if err := newSettings.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = newSettings
	return nil
}

func (s *SettingsService) snapshotSettings() domain.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *SettingsService) GetCSVDelimiters() map[string]rune {
	return map[string]rune{
		"Comma":     ',',
		"Semicolon": ';',
		"Pipe":      '|',
		"Tab":       '\t',
		"Colon":     ':',
	}
}
