package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gingest-desktop/internal/ingest"
	"gingest-desktop/internal/model"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return "Hello " + name
}

func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择要提取的项目目录",
	})
}

func (a *App) ScanLocalDirectory(path string) (model.GingestResponse, error) {
	return ingest.ScanLocalDirectory(path, model.IngestOptions{})
}

func (a *App) SelectAndScanLocalDirectory() (model.GingestResponse, error) {
	path, err := a.SelectDirectory()
	if err != nil {
		return model.GingestResponse{}, err
	}

	if path == "" {
		return model.GingestResponse{}, nil
	}

	return a.ScanLocalDirectory(path)
}
