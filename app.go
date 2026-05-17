package main

import (
	"PriceHike/backend"
	"context"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	Backend *backend.Backend
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{Backend: backend.New()}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Backend.Attach(ctx)

	runtime.OnFileDrop(ctx, func(x, y int, paths []string) {
		if len(paths) == 0 {
			return
		}
		if len(paths) > 1 {
			runtime.EventsEmit(ctx, "file:error", "only one file can be dropped")
			return
		}
		path := paths[0]
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return
		}
		ds, err := a.Backend.LoadFile(path)
		if err != nil {
			runtime.LogErrorf(ctx, "load %s: %v", path, err)
			runtime.EventsEmit(ctx, "file:error", err.Error())
			return
		}
		runtime.EventsEmit(ctx, "file:dropped", filepath.Base(path), ds)
	})
}
