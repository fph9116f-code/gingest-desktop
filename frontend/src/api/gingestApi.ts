import {
    ClearRecentDirectories,
    GetFilterConfig,
    GetRecentDirectories,
    Greet,
    ResetFilterConfig,
    RevealInFileManager,
    SaveFilterConfig,
    SaveXMLFile,
    ScanLocalDirectoryByPath,
    SelectAndScanLocalDirectory,
} from '../../wailsjs/go/main/App'

import type { FilterConfig, GingestResponse, RecentDirectory } from '../types/gingest'

export const gingestApi = {
    greet(name: string): Promise<string> {
        return Greet(name)
    },

    scanLocalProject(): Promise<GingestResponse> {
        return SelectAndScanLocalDirectory() as Promise<GingestResponse>
    },

    scanLocalProjectByPath(path: string): Promise<GingestResponse> {
        return ScanLocalDirectoryByPath(path) as Promise<GingestResponse>
    },

    saveXmlFile(content: string, suggestedFileName: string): Promise<string> {
        return SaveXMLFile(content, suggestedFileName)
    },

    getFilterConfig(): Promise<FilterConfig> {
        return GetFilterConfig() as Promise<FilterConfig>
    },

    saveFilterConfig(config: FilterConfig): Promise<FilterConfig> {
        return SaveFilterConfig(config) as Promise<FilterConfig>
    },

    resetFilterConfig(): Promise<FilterConfig> {
        return ResetFilterConfig() as Promise<FilterConfig>
    },

    getRecentDirectories(): Promise<RecentDirectory[]> {
        return GetRecentDirectories() as Promise<RecentDirectory[]>
    },

    clearRecentDirectories(): Promise<void> {
        return ClearRecentDirectories()
    },
    revealInFileManager(path: string): Promise<void> {
        return RevealInFileManager(path)
    },
}