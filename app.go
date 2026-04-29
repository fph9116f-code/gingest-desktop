package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	appconfig "gingest-desktop/internal/config"
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

func (a *App) GetFilterConfig() (model.FilterConfig, error) {
	return appconfig.LoadFilterConfig()
}

func (a *App) SaveFilterConfig(config model.FilterConfig) (model.FilterConfig, error) {
	config = model.NormalizeFilterConfig(config)
	if err := appconfig.SaveFilterConfig(config); err != nil {
		return model.FilterConfig{}, err
	}
	return config, nil
}

func (a *App) ResetFilterConfig() (model.FilterConfig, error) {
	return appconfig.ResetFilterConfig()
}

func (a *App) ScanLocalDirectory(path string) (model.GingestResponse, error) {
	filterConfig, err := appconfig.LoadFilterConfig()
	if err != nil {
		return model.GingestResponse{}, err
	}

	return ingest.ScanLocalDirectory(path, model.IngestOptions{
		FilterConfig: filterConfig,
	})
}

func (a *App) SelectAndScanLocalDirectory() (model.GingestResponse, error) {
	path, err := a.SelectDirectory()
	if err != nil {
		return model.GingestResponse{}, err
	}

	if path == "" {
		return model.GingestResponse{}, nil
	}

	filterConfig, err := appconfig.LoadFilterConfig()
	if err != nil {
		return model.GingestResponse{}, err
	}

	return ingest.ScanLocalDirectoryWithProgress(path, model.IngestOptions{
		FilterConfig: filterConfig,
	}, func(progress model.ScanProgress) {
		runtime.EventsEmit(a.ctx, "scan-progress", progress)
	})
}

func (a *App) SaveXMLFile(content string, suggestedFileName string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("XML 内容不能为空")
	}

	if strings.TrimSpace(suggestedFileName) == "" {
		suggestedFileName = "gingest_export.xml"
	}

	if !strings.HasSuffix(strings.ToLower(suggestedFileName), ".xml") {
		suggestedFileName += ".xml"
	}

	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存 Gingest XML",
		DefaultFilename: suggestedFileName,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "XML Files (*.xml)",
				Pattern:     "*.xml",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}

	if filePath == "" {
		return "", nil
	}

	if !strings.HasSuffix(strings.ToLower(filePath), ".xml") {
		filePath += ".xml"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filePath, nil
}
