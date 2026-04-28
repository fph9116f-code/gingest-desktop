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
