package services

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DialogService struct {
	ctx context.Context
}

func NewDialogService() *DialogService {
	return &DialogService{}
}

// InitDialogContext wires the runtime context into a DialogService. Declared
// as a package-level function (not a method) so Wails' binding reflection
// doesn't expose it to JS — only methods on the bound struct become callable.
func InitDialogContext(s *DialogService, ctx context.Context) {
	s.ctx = ctx
}

func (s *DialogService) SelectImportFile() (string, error) {
	path, err := runtime.OpenFileDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "Select Import File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Spreadsheets (*.csv, *.xlsx, *.xls)", Pattern: "*.csv;*.xlsx;*.xls"},
			{DisplayName: "CSV (*.csv)", Pattern: "*.csv"},
			{DisplayName: "Excel (*.xlsx, *.xls)", Pattern: "*.xlsx;*.xls"},
			{DisplayName: "Rules (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s *DialogService) SelectExportPath(defaultName, kind string) (string, error) {
	var filter runtime.FileFilter
	switch kind {
	case "csv":
		filter = runtime.FileFilter{DisplayName: "CSV (*.csv)", Pattern: "*.csv"}
	case "xlsx":
		filter = runtime.FileFilter{DisplayName: "Excel (*.xlsx, *.xls)", Pattern: "*.xlsx"}
	case "rules-json":
		filter = runtime.FileFilter{DisplayName: "Rules JSON (*.json)", Pattern: "*.json"}
	default:
		return "", fmt.Errorf("unknown export kind: %s", kind)
	}

	return runtime.SaveFileDialog(s.ctx, runtime.SaveDialogOptions{
		Title:           "Save as",
		DefaultFilename: defaultName,
		Filters:         []runtime.FileFilter{filter},
	})
}
