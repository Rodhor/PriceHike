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

// Attach wires the Wails runtime context into the backend's services. Declared
// as a package-level function (not a method on *Backend) so Wails' binding
// reflection doesn't expose it to JS — only struct methods become callable.
// Call once from app.startup.
func Attach(b *Backend, ctx context.Context) {
	services.InitDialogContext(b.DialogService, ctx)
}
