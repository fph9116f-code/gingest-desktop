<script setup lang="ts">
import {
  Delete,
  FolderOpened,
  Refresh,
} from '@element-plus/icons-vue'
import type { RecentDirectory } from '../types/gingest'

defineProps<{
  directories: RecentDirectory[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'rescan', path: string): void
  (e: 'clear'): void
}>()
</script>

<template>
  <el-card shadow="never" class="recent-card">
    <template #header>
      <div class="card-header">
        <div class="title">
          <el-icon><FolderOpened /></el-icon>
          <strong>最近扫描目录</strong>
        </div>

        <el-button
            v-if="directories.length > 0"
            size="small"
            text
            type="danger"
            :icon="Delete"
            :disabled="loading"
            @click="emit('clear')"
        >
          清空
        </el-button>
      </div>
    </template>

    <el-empty
        v-if="directories.length === 0"
        description="暂无最近扫描目录"
        :image-size="80"
    />

    <div v-else class="recent-list">
      <div
          v-for="item in directories"
          :key="item.path"
          class="recent-item"
      >
        <div class="recent-info">
          <div class="recent-name">
            {{ item.name || item.path }}
          </div>
          <div class="recent-path">
            {{ item.path }}
          </div>
          <div class="recent-time">
            上次扫描：{{ item.lastScanAt || '-' }}
          </div>
        </div>

        <el-button
            size="small"
            type="primary"
            plain
            :icon="Refresh"
            :loading="loading"
            @click="emit('rescan', item.path)"
        >
          重新扫描
        </el-button>
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.recent-card {
  width: min(760px, 100%);
  margin: 20px auto 0;
  text-align: left;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.recent-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #fafafa;
}

.recent-info {
  min-width: 0;
  flex: 1;
}

.recent-name {
  font-weight: 700;
  color: #303133;
  margin-bottom: 4px;
}

.recent-path {
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  color: #606266;
  word-break: break-all;
}

.recent-time {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
</style>