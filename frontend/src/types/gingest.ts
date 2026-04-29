export interface TreeNode {
    id?: number
    label: string
    isFile: boolean
    fullPath?: string
    content?: string
    children?: TreeNode[]
}

export interface FilterConfig {
    ignoreDirectories: string[]
    ignoreExtensions: string[]
    ignoreFileNames: string[]
    maxFileCount: number
    maxTotalSizeMB: number
    maxSingleFileSizeMB: number
}

export interface SkipSample {
    reason: string
    path: string
}

export interface ScanDiagnostics {
    visitedItems: number
    acceptedFiles: number
    skippedItems: number
    skipReasonCounts: Record<string, number>
    skipSamples: SkipSample[]
    noFileHint: string
    effectiveConfig: FilterConfig
    hitFileCountLimit: boolean
    hitTotalSizeLimit: boolean
    stoppedEarly: boolean
    stopReason: string
    stopPath: string
}

export interface GingestResponse {
    projectName: string
    fileCount: number
    estimatedTokens: number
    formattedSize: string
    directoryTree: TreeNode[]
    content: string
    fullContent: string
    diagnostics?: ScanDiagnostics
}

export interface ScanProgress {
    stage: string
    message: string
    currentPath: string
    processedFiles: number
    skippedFiles: number
    totalSize: number
    formattedSize: string
}

export interface FilterConfigForm {
    ignoreDirectoriesText: string
    ignoreExtensionsText: string
    ignoreFileNamesText: string
    maxFileCount: number
    maxTotalSizeMB: number
    maxSingleFileSizeMB: number
}

export interface SkipReasonRow {
    reason: string
    count: number
}

export interface ProjectSummary {
    projectName: string
    fileCount: number
    estimatedTokens: number
    formattedSize: string
}

export interface RecentDirectory {
    path: string
    name: string
    lastScanAt: string
}