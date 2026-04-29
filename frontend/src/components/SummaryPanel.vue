<script setup lang="ts">
import { computed } from 'vue'
import type { FilterConfig, GingestResponse, ProjectSummary, SkipReasonRow } from '../types/gingest'

const props = defineProps<{
  result: GingestResponse
  filterConfig: FilterConfig | null
  message: string
  isSelectedView: boolean
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
      <div class="diagnostics-title">扫描诊断</div>
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

    <div v-if="result.fileCount === 0 && skipReasonRows.length > 0" class="skip-panel">
      <div class="skip-title">跳过原因统计</div>

      <el-table :data="skipReasonRows" size="small" height="160">
        <el-table-column prop="reason" label="原因" min-width="180" />
        <el-table-column prop="count" label="数量" width="80" align="right" />
      </el-table>

      <div
          v-if="result.diagnostics?.skipSamples?.length"
          class="skip-samples"
      >
        <div class="skip-title">跳过样例</div>
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

.diagnostics-box {
  margin-top: 12px;
  padding: 10px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #fafafa;
}

.diagnostics-title,
.skip-title {
  font-size: 13px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 8px;
}

.diagnostics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.diagnostics-grid div {
  padding: 8px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid #ebeef5;
}

.diagnostics-grid span {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.diagnostics-grid strong {
  font-size: 16px;
  color: #409eff;
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