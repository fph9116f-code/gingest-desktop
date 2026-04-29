package ingest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	gitignore "github.com/sabhiram/go-gitignore"

	"gingest-desktop/internal/model"
	"gingest-desktop/internal/utils"
)

func ScanLocalDirectory(rootPath string, options model.IngestOptions) (model.GingestResponse, error) {
	return ScanLocalDirectoryWithProgress(rootPath, options, nil)
}

func ScanLocalDirectoryWithProgress(
	rootPath string,
	options model.IngestOptions,
	onProgress func(model.ScanProgress),
) (model.GingestResponse, error) {
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
	var skippedFiles int64
	var totalSize int64
	var totalTextLength int64
	var lastEmit time.Time

	processedFiles := make([]string, 0)
	fileContents := make(map[string]string)

	emitProgress := func(stage string, message string, currentPath string, force bool) {
		if onProgress == nil {
			return
		}

		if !force && time.Since(lastEmit) < 120*time.Millisecond {
			return
		}

		lastEmit = time.Now()

		onProgress(model.ScanProgress{
			Stage:          stage,
			Message:        message,
			CurrentPath:    currentPath,
			ProcessedFiles: fileCount,
			SkippedFiles:   skippedFiles,
			TotalSize:      totalSize,
			FormattedSize:  utils.FormatSize(totalSize),
		})
	}

	skipFile := func(reason string, currentPath string) {
		skippedFiles++
		emitProgress("scanning", reason, currentPath, false)
	}

	emitProgress("start", "开始扫描本地项目", rootPath, true)

	err = filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			skipFile("跳过无法访问的路径", path)
			return nil
		}

		if path == rootPath {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			skipFile("跳过无法计算相对路径的文件", path)
			return nil
		}

		relativePath = filepath.ToSlash(relativePath)

		if entry.IsDir() {
			if ShouldIgnoreDirectory(entry.Name(), options.FilterConfig) {
				skippedFiles++
				emitProgress("scanning", "跳过忽略目录", relativePath, false)
				return filepath.SkipDir
			}

			if ignoreMatcher != nil && ignoreMatcher.MatchesPath(relativePath+"/") {
				skippedFiles++
				emitProgress("scanning", "根据 .gitignore 跳过目录", relativePath, false)
				return filepath.SkipDir
			}

			emitProgress("scanning", "正在进入目录", relativePath, false)
			return nil
		}

		if ignoreMatcher != nil && ignoreMatcher.MatchesPath(relativePath) {
			skipFile("根据 .gitignore 跳过文件", relativePath)
			return nil
		}

		if ShouldIgnoreFile(entry.Name(), options.FilterConfig) {
			skipFile("跳过忽略文件", relativePath)
			return nil
		}

		fileInfo, err := entry.Info()
		if err != nil {
			skipFile("跳过无法读取信息的文件", relativePath)
			return nil
		}

		if fileInfo.Size() <= 0 {
			skipFile("跳过空文件", relativePath)
			return nil
		}

		if fileInfo.Size() > options.MaxSingleFileSize {
			skipFile("跳过超过单文件大小限制的文件", relativePath)
			return nil
		}

		if fileCount >= options.MaxFileCount {
			return errors.New("安全熔断：代码文件数量超过限制，请缩小扫描范围或调整过滤配置")
		}

		if totalSize+fileInfo.Size() > options.MaxTotalSize {
			return errors.New("安全熔断：源码体积超过限制，请缩小扫描范围或调整过滤配置")
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			skipFile("跳过读取失败的文件", relativePath)
			return nil
		}

		if !utf8.Valid(contentBytes) {
			skipFile("跳过非 UTF-8 文本文件", relativePath)
			return nil
		}

		content := string(contentBytes)

		processedFiles = append(processedFiles, relativePath)
		fileContents[relativePath] = content

		fileCount++
		totalSize += fileInfo.Size()
		totalTextLength += int64(len(content))

		emitProgress("scanning", "正在扫描文件", relativePath, false)

		return nil
	})

	if err != nil {
		emitProgress("error", err.Error(), "", true)
		return model.GingestResponse{}, err
	}

	emitProgress("building", "正在生成目录树", "", true)

	tree := BuildDirectoryTree(processedFiles, fileContents)

	response := model.GingestResponse{
		ProjectName:     "Local: " + filepath.Base(rootPath),
		FileCount:       int(fileCount),
		EstimatedTokens: totalTextLength / 4,
		FormattedSize:   utils.FormatSize(totalSize),
		DirectoryTree:   tree,
	}

	emitProgress("building", "正在组装 XML", "", true)

	fullXML := BuildFullXML(response)
	response.Content = fullXML
	response.FullContent = fullXML

	emitProgress("done", "扫描完成", "", true)

	return response, nil
}

func normalizeOptions(options *model.IngestOptions) {
	if isEmptyFilterConfig(options.FilterConfig) {
		options.FilterConfig = model.DefaultFilterConfig()
	}

	options.FilterConfig = model.NormalizeFilterConfig(options.FilterConfig)

	if options.MaxFileCount <= 0 {
		options.MaxFileCount = options.FilterConfig.MaxFileCount
	}

	if options.MaxTotalSize <= 0 {
		options.MaxTotalSize = options.FilterConfig.MaxTotalSizeMB * 1024 * 1024
	}

	if options.MaxSingleFileSize <= 0 {
		options.MaxSingleFileSize = options.FilterConfig.MaxSingleFileSizeMB * 1024 * 1024
	}
}

func isEmptyFilterConfig(config model.FilterConfig) bool {
	return len(config.IgnoreDirectories) == 0 &&
		len(config.IgnoreExtensions) == 0 &&
		len(config.IgnoreFileNames) == 0 &&
		config.MaxFileCount <= 0 &&
		config.MaxTotalSizeMB <= 0 &&
		config.MaxSingleFileSizeMB <= 0
}
