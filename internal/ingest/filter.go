package ingest

import (
	"path/filepath"
	"strings"
)

var ignoreDirectories = map[string]bool{
	".git":         true,
	".idea":        true,
	".vscode":      true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"out":          true,
	"logs":         true,
	"coverage":     true,
	".next":        true,
	".nuxt":        true,
	".output":      true,
	"vendor":       true,
}

var ignoreFileNames = map[string]bool{
	"package-lock.json": true,
	"pnpm-lock.yaml":    true,
	"yarn.lock":         true,
	"thumbs.db":         true,
	".ds_store":         true,
}

var ignoreExtensions = []string{
	".exe", ".dll", ".so", ".dylib",
	".class", ".jar", ".war", ".ear",
	".zip", ".tar", ".gz", ".rar", ".7z",
	".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".svg",
	".mp3", ".mp4", ".avi", ".mov",
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".ttf", ".otf", ".woff", ".woff2",
	".log",
}

func ShouldIgnoreDirectory(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, ".") {
		return true
	}
	return ignoreDirectories[lower]
}

func ShouldIgnoreFile(name string) bool {
	lower := strings.ToLower(name)

	if ignoreFileNames[lower] {
		return true
	}

	ext := strings.ToLower(filepath.Ext(lower))
	for _, ignoredExt := range ignoreExtensions {
		if ext == ignoredExt {
			return true
		}
	}

	return false
}
