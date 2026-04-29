package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

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
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
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

func (a *App) GetRecentDirectories() ([]model.RecentDirectory, error) {
	return appconfig.LoadRecentDirectories()
}

func (a *App) ClearRecentDirectories() error {
	return appconfig.ClearRecentDirectories()
}

func (a *App) ScanLocalDirectory(path string) (model.GingestResponse, error) {
	return a.scanLocalDirectoryInternal(path, false)
}

func (a *App) ScanLocalDirectoryByPath(path string) (model.GingestResponse, error) {
	return a.scanLocalDirectoryInternal(path, true)
}

func (a *App) SelectAndScanLocalDirectory() (model.GingestResponse, error) {
	path, err := a.SelectDirectory()
	if err != nil {
		return model.GingestResponse{}, err
	}

	if path == "" {
		return model.GingestResponse{}, nil
	}

	return a.scanLocalDirectoryInternal(path, true)
}

func (a *App) scanLocalDirectoryInternal(path string, saveRecent bool) (model.GingestResponse, error) {
	filterConfig, err := appconfig.LoadFilterConfig()
	if err != nil {
		return model.GingestResponse{}, err
	}

	response, err := ingest.ScanLocalDirectoryWithProgress(path, model.IngestOptions{
		FilterConfig: filterConfig,
	}, func(progress model.ScanProgress) {
		wailsruntime.EventsEmit(a.ctx, "scan-progress", progress)
	})

	if err != nil {
		return response, err
	}

	if saveRecent && response.ProjectName != "" {
		_, _ = appconfig.AddRecentDirectory(path)
		wailsruntime.EventsEmit(a.ctx, "recent-directories-changed")
	}

	return response, nil
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

	filePath, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "保存 Gingest XML",
		DefaultFilename: suggestedFileName,
		Filters: []wailsruntime.FileFilter{
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

func (a *App) RevealInFileManager(targetPath string) error {
	targetPath = strings.TrimSpace(targetPath)
	if targetPath == "" {
		return errors.New("文件路径不能为空")
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("文件或目录不存在: %s", absPath)
	}

	if info.IsDir() {
		return openDirectory(absPath)
	}

	return revealFile(absPath)
}

func revealFile(filePath string) error {
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,", filePath).Start()
	case "darwin":
		return exec.Command("open", "-R", filePath).Start()
	default:
		return openDirectory(filepath.Dir(filePath))
	}
}

func openDirectory(dirPath string) error {
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("explorer", dirPath).Start()
	case "darwin":
		return exec.Command("open", dirPath).Start()
	default:
		return exec.Command("xdg-open", dirPath).Start()
	}
}
