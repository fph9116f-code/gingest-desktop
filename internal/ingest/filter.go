package ingest

import (
	"strings"

	"gingest-desktop/internal/model"
)

func ShouldIgnoreDirectory(name string, config model.FilterConfig) bool {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	if lowerName == "" {
		return false
	}

	for _, directory := range config.IgnoreDirectories {
		if lowerName == directory {
			return true
		}
	}

	return false
}

func ShouldIgnoreFile(name string, config model.FilterConfig) bool {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	if lowerName == "" {
		return false
	}

	for _, fileName := range config.IgnoreFileNames {
		if lowerName == fileName {
			return true
		}
	}

	for _, extension := range config.IgnoreExtensions {
		if strings.HasSuffix(lowerName, extension) {
			return true
		}
	}

	return false
}
