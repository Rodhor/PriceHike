package backend

import (
	"PriceHike/backend/services"
	"PriceHike/backend/state"
	"context"
)

type Backend struct {
	*services.DatasetService
	*services.RulesService
	*services.SettingsService
	*services.DialogService

	dialogService *services.DialogService
}

func New() *Backend {
	// Core state
	store := state.New()

	// Services
	dialogSvc := services.NewDialogService()
	settingsSvc := services.NewSettingsService()
	rulesSvc := services.NewRulesService()

	// Dataset service depends on settings and state
	datasetSvc := services.NewDatasetService(settingsSvc, store)

	return &Backend{
		DatasetService:  datasetSvc,
		RulesService:    rulesSvc,
		SettingsService: settingsSvc,
		DialogService:   dialogSvc,
	}
}

// Attach attaches the backend to the given context.
// This is not called from the frontend and should only be called once.
func (b *Backend) Attach(ctx context.Context) {
	b.DialogService.AttachDialogContext(ctx)
}
