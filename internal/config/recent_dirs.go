package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"gingest-desktop/internal/model"
)

const recentDirectoriesFileName = "recent_directories.json"
const maxRecentDirectories = 10

func LoadRecentDirectories() ([]model.RecentDirectory, error) {
	configPath, err := getRecentDirectoriesPath()
	if err != nil {
		return []model.RecentDirectory{}, err
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return []model.RecentDirectory{}, nil
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return []model.RecentDirectory{}, err
	}

	var recentDirectories []model.RecentDirectory
	if err := json.Unmarshal(bytes, &recentDirectories); err != nil {
		return []model.RecentDirectory{}, err
	}

	return recentDirectories, nil
}

func AddRecentDirectory(path string) ([]model.RecentDirectory, error) {
	if path == "" {
		return LoadRecentDirectories()
	}

	info, err := os.Stat(path)
	if err != nil {
		return LoadRecentDirectories()
	}

	if !info.IsDir() {
		return LoadRecentDirectories()
	}

	current := model.RecentDirectory{
		Path:       path,
		Name:       filepath.Base(path),
		LastScanAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	recentDirectories, err := LoadRecentDirectories()
	if err != nil {
		recentDirectories = []model.RecentDirectory{}
	}

	deduped := make([]model.RecentDirectory, 0, maxRecentDirectories)
	deduped = append(deduped, current)

	for _, item := range recentDirectories {
		if item.Path == path {
			continue
		}

		deduped = append(deduped, item)

		if len(deduped) >= maxRecentDirectories {
			break
		}
	}

	if err := SaveRecentDirectories(deduped); err != nil {
		return deduped, err
	}

	return deduped, nil
}

func SaveRecentDirectories(recentDirectories []model.RecentDirectory) error {
	configPath, err := getRecentDirectoriesPath()
	if err != nil {
		return err
	}

	if len(recentDirectories) > maxRecentDirectories {
		recentDirectories = recentDirectories[:maxRecentDirectories]
	}

	bytes, err := json.MarshalIndent(recentDirectories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, bytes, 0644)
}

func ClearRecentDirectories() error {
	configPath, err := getRecentDirectoriesPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return os.Remove(configPath)
}

func getRecentDirectoriesPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(baseDir, appConfigDirName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, recentDirectoriesFileName), nil
}
