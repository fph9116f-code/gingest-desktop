<script setup lang="ts">
import { computed } from 'vue'
import type { ScanProgress } from '../types/gingest'

const props = defineProps<{
  loading: boolean
  progress: ScanProgress | null
}>()

const percentage = computed(() => {
  if (!props.loading) return 100
  if (!props.progress) return 5

  switch (props.progress.stage) {
    case 'start':
      return 10
    case 'scanning':
      return 45
    case 'building':
      return 80
    case 'limit':
      return 95
    case 'done':
      return 100
    case 'error':
      return 100
    default:
      return 20
  }
})
</script>

<template>
  <el-card v-if="loading && progress" shadow="never" class="progress-card">
    <div class="progress-header">
      <strong>扫描进度</strong>
      <el-tag type="info">
        {{ progress.stage }}
      </el-tag>
    </div>

    <el-progress
        :percentage="percentage"
        :indeterminate="progress.stage === 'scanning'"
    />

    <div class="progress-detail">
      <div>
        <strong>状态：</strong>
        {{ progress.message || '准备中...' }}
      </div>

      <div>
        <strong>当前：</strong>
        <span class="current-path">{{ progress.currentPath || '-' }}</span>
      </div>

      <div class="progress-stats">
        <span>已处理：{{ progress.processedFiles || 0 }} 文件</span>
        <span>已跳过：{{ progress.skippedFiles || 0 }} 项</span>
        <span>已读取：{{ progress.formattedSize || '0 B' }}</span>
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.progress-card {
  flex-shrink: 0;
}

.progress-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.progress-detail {
  margin-top: 12px;
  font-size: 13px;
  color: #606266;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.current-path {
  font-family: Consolas, "Courier New", monospace;
  word-break: break-all;
}

.progress-stats {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}
</style>