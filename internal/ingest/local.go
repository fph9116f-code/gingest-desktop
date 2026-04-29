package ingest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	gitignore "github.com/sabhiram/go-gitignore"

	"gingest-desktop/internal/model"
	"gingest-desktop/internal/utils"
)

const maxSkipSamples = 20

var errStopScanByLimit = errors.New("scan stopped by configured limit")

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

	var visitedItems int64
	var fileCount int64
	var skippedFiles int64
	var totalSize int64
	var totalTextLength int64
	var lastEmit time.Time
	var hitFileCountLimit bool
	var hitTotalSizeLimit bool
	var stoppedEarly bool
	var stopReason string
	var stopPath string

	skipReasonCounts := make(map[string]int64)
	skipSamples := make([]model.SkipSample, 0, maxSkipSamples)

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

	addSkip := func(reason string, currentPath string) {
		skippedFiles++
		skipReasonCounts[reason]++

		if len(skipSamples) < maxSkipSamples {
			skipSamples = append(skipSamples, model.SkipSample{
				Reason: reason,
				Path:   currentPath,
			})
		}

		emitProgress("scanning", reason, currentPath, false)
	}

	emitProgress("start", "开始扫描本地项目", rootPath, true)

	err = filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			addSkip("无法访问路径", path)
			return nil
		}

		if path == rootPath {
			return nil
		}

		visitedItems++

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			addSkip("无法计算相对路径", path)
			return nil
		}

		relativePath = filepath.ToSlash(relativePath)

		if entry.IsDir() {
			if ShouldIgnoreDirectory(entry.Name(), options.FilterConfig) {
				addSkip("命中过滤目录", relativePath)
				return filepath.SkipDir
			}

			if ignoreMatcher != nil && ignoreMatcher.MatchesPath(relativePath+"/") {
				addSkip("命中 .gitignore 目录规则", relativePath)
				return filepath.SkipDir
			}

			emitProgress("scanning", "正在进入目录", relativePath, false)
			return nil
		}

		if ignoreMatcher != nil && ignoreMatcher.MatchesPath(relativePath) {
			addSkip("命中 .gitignore 文件规则", relativePath)
			return nil
		}

		if ShouldIgnoreFile(entry.Name(), options.FilterConfig) {
			addSkip("命中过滤文件名或扩展名", relativePath)
			return nil
		}

		fileInfo, err := entry.Info()
		if err != nil {
			addSkip("无法读取文件信息", relativePath)
			return nil
		}

		if fileInfo.Size() <= 0 {
			addSkip("空文件", relativePath)
			return nil
		}

		if fileInfo.Size() > options.MaxSingleFileSize {
			addSkip(fmt.Sprintf("超过单文件大小限制：%s > %s", utils.FormatSize(fileInfo.Size()), utils.FormatSize(options.MaxSingleFileSize)), relativePath)
			return nil
		}

		if fileCount >= options.MaxFileCount {
			hitFileCountLimit = true
			stoppedEarly = true
			stopPath = relativePath
			stopReason = fmt.Sprintf(
				"已达到最大文件数限制：当前有效文件 %d 个，配置上限 %d 个。扫描已提前停止。请调大「最大文件数」或缩小扫描范围。",
				fileCount,
				options.MaxFileCount,
			)
			emitProgress("limit", stopReason, relativePath, true)
			return errStopScanByLimit
		}

		if totalSize+fileInfo.Size() > options.MaxTotalSize {
			hitTotalSizeLimit = true
			stoppedEarly = true
			stopPath = relativePath
			stopReason = fmt.Sprintf(
				"已达到总大小限制：当前已读取 %s，即将读取文件 %s，配置上限 %s。扫描已提前停止。请调大「总大小上限 MB」或缩小扫描范围。",
				utils.FormatSize(totalSize),
				utils.FormatSize(fileInfo.Size()),
				utils.FormatSize(options.MaxTotalSize),
			)
			emitProgress("limit", stopReason, relativePath, true)
			return errStopScanByLimit
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			addSkip("读取文件失败", relativePath)
			return nil
		}

		if !utf8.Valid(contentBytes) {
			addSkip("非 UTF-8 文本文件", relativePath)
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

	diagnostics := model.ScanDiagnostics{
		VisitedItems:      visitedItems,
		AcceptedFiles:     fileCount,
		SkippedItems:      skippedFiles,
		SkipReasonCounts:  skipReasonCounts,
		SkipSamples:       skipSamples,
		EffectiveConfig:   options.FilterConfig,
		HitFileCountLimit: hitFileCountLimit,
		HitTotalSizeLimit: hitTotalSizeLimit,
		StoppedEarly:      stoppedEarly,
		StopReason:        stopReason,
		StopPath:          stopPath,
	}

	if err != nil && !errors.Is(err, errStopScanByLimit) {
		emitProgress("error", err.Error(), "", true)
		return model.GingestResponse{
			Diagnostics: diagnostics,
		}, err
	}

	emitProgress("building", "正在生成目录树", "", true)

	tree := BuildDirectoryTree(processedFiles, fileContents)

	response := model.GingestResponse{
		ProjectName:     "Local: " + filepath.Base(rootPath),
		FileCount:       int(fileCount),
		EstimatedTokens: totalTextLength / 4,
		FormattedSize:   utils.FormatSize(totalSize),
		DirectoryTree:   tree,
		Diagnostics:     diagnostics,
	}

	if fileCount == 0 {
		response.Diagnostics.NoFileHint = buildNoFileHint(diagnostics)
		response.Content = buildEmptyScanContent(response)
		response.FullContent = response.Content
		emitProgress("done", "扫描完成，但没有匹配到有效文件", "", true)
		return response, nil
	}

	emitProgress("building", "正在组装 XML", "", true)

	fullXML := BuildFullXML(response)
	response.Content = fullXML
	response.FullContent = fullXML

	emitProgress("done", "扫描完成", "", true)

	return response, nil
}

func buildNoFileHint(diagnostics model.ScanDiagnostics) string {
	if diagnostics.VisitedItems == 0 {
		return "没有扫描到任何子目录或文件。请确认选择的是项目根目录，而不是空目录。"
	}

	if diagnostics.SkippedItems == 0 {
		return "扫描到了路径，但没有发现可读取的文本代码文件。请检查目录内容。"
	}

	topReason := getTopSkipReason(diagnostics.SkipReasonCounts)

	switch topReason {
	case "命中过滤目录":
		return "大部分内容被「忽略目录」规则过滤了。请检查过滤配置里的目录名，例如 src、frontend、internal 是否被误填。"
	case "命中过滤文件名或扩展名":
		return "大部分文件被「忽略文件名或扩展名」规则过滤了。请检查是否误加入了 .go、.vue、.ts、.java、.xml、.yml 等代码扩展名。"
	case "命中 .gitignore 目录规则", "命中 .gitignore 文件规则":
		return "文件主要被项目 .gitignore 规则过滤了。请检查 .gitignore 是否包含了当前需要提取的目录或文件。"
	case "非 UTF-8 文本文件":
		return "文件主要因为不是 UTF-8 编码而被跳过。可以后续增加 GBK/其他编码识别。"
	default:
		if strings.HasPrefix(topReason, "超过单文件大小限制") {
			return "文件主要因为超过单文件大小限制被跳过。请在过滤配置里调大「单文件上限 MB」。"
		}
		return "没有匹配到有效文件。请查看下方跳过原因统计和样例，调整过滤配置后重新扫描。"
	}
}

func getTopSkipReason(reasonCounts map[string]int64) string {
	type pair struct {
		reason string
		count  int64
	}

	pairs := make([]pair, 0, len(reasonCounts))
	for reason, count := range reasonCounts {
		pairs = append(pairs, pair{reason: reason, count: count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	if len(pairs) == 0 {
		return ""
	}

	return pairs[0].reason
}

func buildEmptyScanContent(response model.GingestResponse) string {
	var sb strings.Builder

	sb.WriteString("<project_summary>\n")
	sb.WriteString("Project: ")
	sb.WriteString(response.ProjectName)
	sb.WriteString("\n")
	sb.WriteString("Total Files: 0\n")
	sb.WriteString("Estimated Tokens: 0\n")
	sb.WriteString("</project_summary>\n\n")

	sb.WriteString("<scan_diagnostics>\n")
	sb.WriteString("Hint: ")
	sb.WriteString(response.Diagnostics.NoFileHint)
	sb.WriteString("\n")
	sb.WriteString("Visited Items: ")
	sb.WriteString(fmt.Sprintf("%d", response.Diagnostics.VisitedItems))
	sb.WriteString("\n")
	sb.WriteString("Skipped Items: ")
	sb.WriteString(fmt.Sprintf("%d", response.Diagnostics.SkippedItems))
	sb.WriteString("\n")
	sb.WriteString("</scan_diagnostics>\n\n")

	sb.WriteString("<files>\n</files>")

	return sb.String()
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
