package model

type GingestResponse struct {
	ProjectName     string     `json:"projectName"`
	FileCount       int        `json:"fileCount"`
	EstimatedTokens int64      `json:"estimatedTokens"`
	FormattedSize   string     `json:"formattedSize"`
	DirectoryTree   []TreeNode `json:"directoryTree"`
	Content         string     `json:"content"`
	FullContent     string     `json:"fullContent"`
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
