package services

import (
	"PriceHike/backend/domain"
	"PriceHike/backend/engine"
	csvio "PriceHike/backend/io/csv"
	excelio "PriceHike/backend/io/excel"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SettingsProvider interface {
	snapshotSettings() domain.Settings
}

type DataStore interface {
	Get() (domain.Dataset, error)
	Set(ds domain.Dataset, path string)
}

type DatasetService struct {
	settingsProvider SettingsProvider
	store            DataStore
}

func NewDatasetService(settingsProvider SettingsProvider, store DataStore) *DatasetService {
	return &DatasetService{
		settingsProvider: settingsProvider,
		store:            store,
	}
}

func (s *DatasetService) LoadFile(path string) (domain.Dataset, error) {
	settings := s.settingsProvider.snapshotSettings()
	var (
		ds  *domain.Dataset
		err error
	)
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xlsx", ".xls":
		ds, err = excelio.Read(path)
	case ".csv":
		ds, err = csvio.Read(path, settings.Delimiter())
	default:
		return domain.Dataset{}, domain.ErrUnsupportedFormat
	}
	if err != nil {
		return domain.Dataset{}, err
	}
	s.store.Set(*ds, path)
	return *ds, nil
}

func (s *DatasetService) GeneratePreview(rules []domain.Rule, sampleSize int) (domain.PreviewResult, error) {
	ds, err := s.store.Get()
	if err != nil {
		return domain.PreviewResult{}, err
	}
	return engine.Preview(rules, ds, sampleSize, s.settingsProvider.snapshotSettings().DefaultDecimalPlaces)
}

func (s *DatasetService) ApplyAndExportCSV(rules []domain.Rule, destPath string) (domain.ExportReport, error) {
	return s.exportWith(rules, destPath, func(ds domain.Dataset) error {
		return csvio.Write(destPath, ds, s.settingsProvider.snapshotSettings().Delimiter())
	})

}

func (s *DatasetService) ApplyAndExportXLSX(rules []domain.Rule, destPath string) (domain.ExportReport, error) {
	return s.exportWith(rules, destPath, func(ds domain.Dataset) error {
		return excelio.Write(destPath, ds)
	})
}

func (s *DatasetService) exportWith(
	rules []domain.Rule,
	destPath string,
	write func(domain.Dataset) error,
) (domain.ExportReport, error) {
	start := time.Now().UTC()
	ds, err := s.store.Get()
	if err != nil {
		return domain.ExportReport{}, err
	}
	out, _, err := engine.ApplyRules(rules, ds, s.settingsProvider.snapshotSettings().DefaultDecimalPlaces)
	if err != nil {
		return domain.ExportReport{}, err
	}
	if err := write(out); err != nil {
		return domain.ExportReport{}, fmt.Errorf("write failed: %w", err)
	}
	info, _ := os.Stat(destPath)
	return domain.ExportReport{
		Path:           destPath,
		ItemsWritten:   len(out.Items),
		BytesWritten:   info.Size(),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
