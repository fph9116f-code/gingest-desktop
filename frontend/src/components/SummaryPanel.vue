<script setup lang="ts">
import { computed } from 'vue'
import type {
  FilterConfig,
  GingestResponse,
  ProjectSummary,
  SelectedFileStats,
  SkipReasonRow,
  TreeNode,
} from '../types/gingest'
import { formatToken, getTokenLevel } from '../utils/format'

const props = defineProps<{
  result: GingestResponse
  filterConfig: FilterConfig | null
  message: string
  isSelectedView: boolean
  selectedStats?: SelectedFileStats
}>()

const summary = computed<ProjectSummary>(() => ({
  projectName: props.result.projectName,
  fileCount: props.result.fileCount,
  estimatedTokens: props.result.estimatedTokens,
  formattedSize: props.result.formattedSize,
}))

const skipReasonRows = computed<SkipReasonRow[]>(() => {
  const counts = props.result.diagnostics?.skipReasonCounts || {}

  return Object.entries(counts)
      .map(([reason, count]) => ({
        reason,
        count,
      }))
      .sort((a, b) => b.count - a.count)
})

const flattenFiles = (nodes: TreeNode[]) => {
  const result: TreeNode[] = []

  const walk = (items: TreeNode[]) => {
    items.forEach((item) => {
      if (item.isFile) {
        result.push(item)
        return
      }

      if (item.children?.length) {
        walk(item.children)
      }
    })
  }

  walk(nodes)

  return result
}

const largeFiles = computed(() => {
  return flattenFiles(props.result.directoryTree || [])
      .sort((a, b) => (b.estimatedTokens || 0) - (a.estimatedTokens || 0))
      .slice(0, 10)
})

const selectedStats = computed(() => {
  return props.selectedStats || {
    fileCount: 0,
    sizeBytes: 0,
    formattedSize: '0 B',
    estimatedTokens: 0,
  }
})

const getTokenTagType = (tokens: number) => {
  const level = getTokenLevel(tokens)

  if (level === 'danger') return 'danger'
  if (level === 'warning') return 'warning'
  return 'info'
}
</script>

<template>
  <el-card shadow="never" class="summary-card">
    <template #header>
      <div class="card-header">
        <strong>项目摘要</strong>
        <el-tag v-if="isSelectedView" type="warning">局部视图</el-tag>
        <el-tag v-else type="success">全库视图</el-tag>
      </div>
    </template>

    <el-descriptions :column="1" border>
      <el-descriptions-item label="项目">
        {{ summary.projectName }}
      </el-descriptions-item>
      <el-descriptions-item label="文件数">
        {{ summary.fileCount }}
      </el-descriptions-item>
      <el-descriptions-item label="Tokens">
        {{ summary.estimatedTokens }}
      </el-descriptions-item>
      <el-descriptions-item label="大小">
        {{ summary.formattedSize }}
      </el-descriptions-item>
      <el-descriptions-item label="Go 返回">
        {{ message || '暂无' }}
      </el-descriptions-item>
    </el-descriptions>

    <div class="selected-box">
      <div class="section-title">勾选预算</div>

      <div class="selected-grid">
        <div>
          <span>文件</span>
          <strong>{{ selectedStats.fileCount }}</strong>
        </div>

        <div>
          <span>Tokens</span>
          <strong>{{ formatToken(selectedStats.estimatedTokens) }}</strong>
        </div>

        <div>
          <span>大小</span>
          <strong>{{ selectedStats.formattedSize }}</strong>
        </div>
      </div>
    </div>

    <div class="summary-tip">
      当前过滤规则：
      <span>{{ filterConfig?.ignoreDirectories?.length || 0 }}</span> 个目录，
      <span>{{ filterConfig?.ignoreExtensions?.length || 0 }}</span> 个扩展名，
      <span>{{ filterConfig?.ignoreFileNames?.length || 0 }}</span> 个文件名。
    </div>

    <el-alert
        v-if="result.diagnostics?.stoppedEarly"
        :title="result.diagnostics.stopReason"
        type="warning"
        show-icon
        :closable="false"
        class="diagnostics-alert"
    />

    <el-alert
        v-if="result.fileCount === 0 && result.diagnostics?.noFileHint"
        :title="result.diagnostics.noFileHint"
        type="warning"
        show-icon
        :closable="false"
        class="diagnostics-alert"
    />

    <div v-if="result.diagnostics" class="diagnostics-box">
      <div class="section-title">扫描诊断</div>
      <div class="diagnostics-grid">
        <div>
          <span>访问项</span>
          <strong>{{ result.diagnostics.visitedItems }}</strong>
        </div>
        <div>
          <span>有效文件</span>
          <strong>{{ result.diagnostics.acceptedFiles }}</strong>
        </div>
        <div>
          <span>跳过项</span>
          <strong>{{ result.diagnostics.skippedItems }}</strong>
        </div>
      </div>
    </div>

    <div v-if="largeFiles.length > 0" class="large-files-box">
      <div class="section-title">大文件 Top 10</div>

      <div class="large-file-list">
        <div
            v-for="file in largeFiles"
            :key="file.fullPath || file.label"
            class="large-file-item"
        >
          <div class="large-file-path">
            {{ file.fullPath || file.label }}
          </div>

          <div class="large-file-tags">
            <el-tag size="small" type="info" effect="plain">
              {{ file.formattedSize || '0 B' }}
            </el-tag>

            <el-tag
                size="small"
                :type="getTokenTagType(file.estimatedTokens || 0)"
                effect="plain"
            >
              {{ formatToken(file.estimatedTokens || 0) }} tokens
            </el-tag>
          </div>
        </div>
      </div>
    </div>

    <div v-if="result.fileCount === 0 && skipReasonRows.length > 0" class="skip-panel">
      <div class="section-title">跳过原因统计</div>

      <el-table :data="skipReasonRows" size="small" height="160">
        <el-table-column prop="reason" label="原因" min-width="180" />
        <el-table-column prop="count" label="数量" width="80" align="right" />
      </el-table>

      <div
          v-if="result.diagnostics?.skipSamples?.length"
          class="skip-samples"
      >
        <div class="section-title">跳过样例</div>
        <div
            v-for="item in result.diagnostics.skipSamples"
            :key="item.reason + item.path"
            class="skip-sample-item"
        >
          <el-tag size="small" type="info">{{ item.reason }}</el-tag>
          <span>{{ item.path }}</span>
        </div>
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.summary-card {
  overflow: hidden;
}

.summary-card :deep(.el-card__body) {
  height: calc(100% - 58px);
  overflow: auto;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.summary-tip {
  margin-top: 12px;
  font-size: 12px;
  color: #909399;
  line-height: 1.6;
}

.summary-tip span {
  color: #409eff;
  font-weight: 700;
}

.diagnostics-alert {
  margin-top: 12px;
}

.selected-box,
.diagnostics-box,
.large-files-box {
  margin-top: 12px;
  padding: 10px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #fafafa;
}

.section-title {
  font-size: 13px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 8px;
}

.selected-grid,
.diagnostics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.selected-grid div,
.diagnostics-grid div {
  padding: 8px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid #ebeef5;
}

.selected-grid span,
.diagnostics-grid span {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.selected-grid strong,
.diagnostics-grid strong {
  font-size: 16px;
  color: #409eff;
}

.large-file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.large-file-item {
  padding: 8px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid #ebeef5;
}

.large-file-path {
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  color: #606266;
  word-break: break-all;
  margin-bottom: 6px;
}

.large-file-tags {
  display: flex;
  align-items: center;
  gap: 6px;
}

.skip-panel {
  margin-top: 12px;
  border-top: 1px solid #ebeef5;
  padding-top: 12px;
}

.skip-samples {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.skip-sample-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  color: #606266;
  word-break: break-all;
}
</style>