<script setup lang="ts">
import {
  Connection,
  FolderOpened,
  Refresh,
  Setting,
} from '@element-plus/icons-vue'

defineProps<{
  loading: boolean
  hasResult: boolean
}>()

const emit = defineEmits<{
  (e: 'test-go'): void
  (e: 'open-settings'): void
  (e: 'scan-local'): void
  (e: 'clear-result'): void
}>()
</script>

<template>
  <el-header class="header">
    <div class="title">
      <span>Gingest Desktop</span>
      <el-tag type="success">Wails + Go + Vue3 + TS</el-tag>
    </div>

    <div class="actions">
      <el-button :icon="Connection" @click="emit('test-go')">
        测试 Go
      </el-button>

      <el-button :icon="Setting" @click="emit('open-settings')">
        过滤配置
      </el-button>

      <el-button
          type="primary"
          :icon="FolderOpened"
          :loading="loading"
          @click="emit('scan-local')"
      >
        选择并扫描本地项目
      </el-button>

      <el-button
          v-if="hasResult"
          type="info"
          plain
          :icon="Refresh"
          @click="emit('clear-result')"
      >
        清空
      </el-button>
    </div>
  </el-header>
</template>

<style scoped>
.header {
  height: 64px;
  background: #24292f;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 700;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>