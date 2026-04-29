package model

import "strings"

type GingestResponse struct {
	ProjectName     string          `json:"projectName"`
	FileCount       int             `json:"fileCount"`
	EstimatedTokens int64           `json:"estimatedTokens"`
	FormattedSize   string          `json:"formattedSize"`
	DirectoryTree   []TreeNode      `json:"directoryTree"`
	Content         string          `json:"content"`
	FullContent     string          `json:"fullContent"`
	Diagnostics     ScanDiagnostics `json:"diagnostics"`
}

type TreeNode struct {
	ID       int        `json:"id,omitempty"`
	Label    string     `json:"label"`
	IsFile   bool       `json:"isFile"`
	FullPath string     `json:"fullPath,omitempty"`
	Content  string     `json:"content,omitempty"`
	Children []TreeNode `json:"children,omitempty"`
}

type IngestOptions struct {
	MaxFileCount      int64
	MaxTotalSize      int64
	MaxSingleFileSize int64
	FilterConfig      FilterConfig
}

type ScanProgress struct {
	Stage          string `json:"stage"`
	Message        string `json:"message"`
	CurrentPath    string `json:"currentPath"`
	ProcessedFiles int64  `json:"processedFiles"`
	SkippedFiles   int64  `json:"skippedFiles"`
	TotalSize      int64  `json:"totalSize"`
	FormattedSize  string `json:"formattedSize"`
}

type SkipSample struct {
	Reason string `json:"reason"`
	Path   string `json:"path"`
}

type ScanDiagnostics struct {
	VisitedItems      int64            `json:"visitedItems"`
	AcceptedFiles     int64            `json:"acceptedFiles"`
	SkippedItems      int64            `json:"skippedItems"`
	SkipReasonCounts  map[string]int64 `json:"skipReasonCounts"`
	SkipSamples       []SkipSample     `json:"skipSamples"`
	NoFileHint        string           `json:"noFileHint"`
	EffectiveConfig   FilterConfig     `json:"effectiveConfig"`
	HitFileCountLimit bool             `json:"hitFileCountLimit"`
	HitTotalSizeLimit bool             `json:"hitTotalSizeLimit"`

	// 新增：扫描是否被配置限制提前停止
	StoppedEarly bool   `json:"stoppedEarly"`
	StopReason   string `json:"stopReason"`
	StopPath     string `json:"stopPath"`
}

type FilterConfig struct {
	IgnoreDirectories   []string `json:"ignoreDirectories"`
	IgnoreExtensions    []string `json:"ignoreExtensions"`
	IgnoreFileNames     []string `json:"ignoreFileNames"`
	MaxFileCount        int64    `json:"maxFileCount"`
	MaxTotalSizeMB      int64    `json:"maxTotalSizeMB"`
	MaxSingleFileSizeMB int64    `json:"maxSingleFileSizeMB"`
}

type RecentDirectory struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	LastScanAt string `json:"lastScanAt"`
}

func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		IgnoreDirectories: []string{
			".git",
			".idea",
			".vscode",
			"node_modules",
			"dist",
			"build",
			"target",
			"out",
			"logs",
			"coverage",
			".next",
			".nuxt",
			".output",
			"vendor",
		},
		IgnoreExtensions: []string{
			".exe",
			".dll",
			".so",
			".dylib",
			".class",
			".jar",
			".war",
			".ear",
			".zip",
			".tar",
			".gz",
			".rar",
			".7z",
			".png",
			".jpg",
			".jpeg",
			".gif",
			".webp",
			".ico",
			".svg",
			".mp3",
			".mp4",
			".avi",
			".mov",
			".pdf",
			".doc",
			".docx",
			".xls",
			".xlsx",
			".ppt",
			".pptx",
			".ttf",
			".otf",
			".woff",
			".woff2",
			".log",
			".min.js",
			".min.css",
			".map",
		},
		IgnoreFileNames: []string{
			"package-lock.json",
			"pnpm-lock.yaml",
			"yarn.lock",
			"thumbs.db",
			".ds_store",
		},
		MaxFileCount:        3000,
		MaxTotalSizeMB:      50,
		MaxSingleFileSizeMB: 2,
	}
}

func NormalizeFilterConfig(config FilterConfig) FilterConfig {
	defaultConfig := DefaultFilterConfig()

	if config.MaxFileCount <= 0 {
		config.MaxFileCount = defaultConfig.MaxFileCount
	}

	if config.MaxTotalSizeMB <= 0 {
		config.MaxTotalSizeMB = defaultConfig.MaxTotalSizeMB
	}

	if config.MaxSingleFileSizeMB <= 0 {
		config.MaxSingleFileSizeMB = defaultConfig.MaxSingleFileSizeMB
	}

	config.IgnoreDirectories = normalizeStringList(config.IgnoreDirectories, false)
	config.IgnoreExtensions = normalizeStringList(config.IgnoreExtensions, true)
	config.IgnoreFileNames = normalizeStringList(config.IgnoreFileNames, false)

	return config
}

func normalizeStringList(values []string, isExtension bool) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool)

	for _, value := range values {
		item := strings.TrimSpace(strings.ToLower(value))
		if item == "" {
			continue
		}

		if isExtension && !strings.HasPrefix(item, ".") {
			item = "." + item
		}

		if seen[item] {
			continue
		}

		seen[item] = true
		result = append(result, item)
	}

	return result
}
