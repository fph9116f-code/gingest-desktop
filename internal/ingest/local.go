package ingest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	gitignore "github.com/sabhiram/go-gitignore"

	"gingest-desktop/internal/model"
	"gingest-desktop/internal/utils"
)

const (
	defaultMaxFileCount      int64 = 3000
	defaultMaxTotalSize      int64 = 50 * 1024 * 1024
	defaultMaxSingleFileSize int64 = 2 * 1024 * 1024
)

func ScanLocalDirectory(rootPath string, options model.IngestOptions) (model.GingestResponse, error) {
	if strings.TrimSpace(rootPath) == "" {
		return model.GingestResponse{}, errors.New("目录路径不能为空")
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return model.GingestResponse{}, err
	}

	if !info.IsDir() {
		return model.GingestResponse{}, errors.New("请选择一个有效目录")
	}

	normalizeOptions(&options)

	var ignoreMatcher *gitignore.GitIgnore
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		ignoreMatcher, _ = gitignore.CompileIgnoreFile(gitignorePath)
	}

	var fileCount int64
	var totalSize int64
	var totalTextLength int64

	processedFiles := make([]string, 0)
	fileContents := make(map[string]string)

	err = filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if path == rootPath {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil
		}

		relativePath = filepath.ToSlash(relativePath)

		if entry.IsDir() {
			if ShouldIgnoreDirectory(entry.Name()) {
				return filepath.SkipDir
			}

			if ignoreMatcher != nil && ignoreMatcher.MatchesPath(relativePath+"/") {
				return filepath.SkipDir
			}

			return nil
		}

		if ignoreMatcher != nil && ignoreMatcher.MatchesPath(relativePath) {
			return nil
		}

		if ShouldIgnoreFile(entry.Name()) {
			return nil
		}

		fileInfo, err := entry.Info()
		if err != nil {
			return nil
		}

		if fileInfo.Size() <= 0 {
			return nil
		}

		if fileInfo.Size() > options.MaxSingleFileSize {
			return nil
		}

		if fileCount >= options.MaxFileCount {
			return errors.New("安全熔断：代码文件数量超过限制，请缩小扫描范围")
		}

		if totalSize+fileInfo.Size() > options.MaxTotalSize {
			return errors.New("安全熔断：源码体积超过限制，请缩小扫描范围")
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if !utf8.Valid(contentBytes) {
			return nil
		}

		content := string(contentBytes)

		processedFiles = append(processedFiles, relativePath)
		fileContents[relativePath] = content

		fileCount++
		totalSize += fileInfo.Size()
		totalTextLength += int64(len(content))

		return nil
	})

	if err != nil {
		return model.GingestResponse{}, err
	}

	tree := BuildDirectoryTree(processedFiles, fileContents)

	response := model.GingestResponse{
		ProjectName:     "Local: " + filepath.Base(rootPath),
		FileCount:       int(fileCount),
		EstimatedTokens: totalTextLength / 4,
		FormattedSize:   utils.FormatSize(totalSize),
		DirectoryTree:   tree,
	}

	fullXML := BuildFullXML(response)
	response.Content = fullXML
	response.FullContent = fullXML

	return response, nil
}

func normalizeOptions(options *model.IngestOptions) {
	if options.MaxFileCount <= 0 {
		options.MaxFileCount = defaultMaxFileCount
	}

	if options.MaxTotalSize <= 0 {
		options.MaxTotalSize = defaultMaxTotalSize
	}

	if options.MaxSingleFileSize <= 0 {
		options.MaxSingleFileSize = defaultMaxSingleFileSize
	}
}
