package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"gingest-desktop/internal/model"
)

const appConfigDirName = "GingestDesktop"
const filterConfigFileName = "filter_config.json"

func LoadFilterConfig() (model.FilterConfig, error) {
	configPath, err := getFilterConfigPath()
	if err != nil {
		return model.DefaultFilterConfig(), err
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		defaultConfig := model.DefaultFilterConfig()
		if saveErr := SaveFilterConfig(defaultConfig); saveErr != nil {
			return defaultConfig, nil
		}
		return defaultConfig, nil
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return model.DefaultFilterConfig(), err
	}

	var config model.FilterConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return model.DefaultFilterConfig(), err
	}

	return model.NormalizeFilterConfig(config), nil
}

func SaveFilterConfig(config model.FilterConfig) error {
	configPath, err := getFilterConfigPath()
	if err != nil {
		return err
	}

	config = model.NormalizeFilterConfig(config)

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, bytes, 0644)
}

func ResetFilterConfig() (model.FilterConfig, error) {
	defaultConfig := model.DefaultFilterConfig()
	if err := SaveFilterConfig(defaultConfig); err != nil {
		return defaultConfig, err
	}
	return defaultConfig, nil
}

func getFilterConfigPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(baseDir, appConfigDirName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, filterConfigFileName), nil
}
