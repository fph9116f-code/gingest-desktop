<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { FolderOpened, Connection } from '@element-plus/icons-vue'

import { Greet, SelectDirectory } from '../wailsjs/go/main/App'

const message = ref('')
const selectedPath = ref('')

const testGoCall = async () => {
  try {
    message.value = await Greet('Gingest Desktop')
    ElMessage.success('Go 调用成功')
  } catch (error) {
    console.error(error)
    ElMessage.error('Go 调用失败')
  }
}

const handleSelectDirectory = async () => {
  try {
    const path = await SelectDirectory()
    if (!path) return

    selectedPath.value = path
    ElMessage.success('目录选择成功')
  } catch (error) {
    console.error(error)
    ElMessage.error('选择目录失败')
  }
}
</script>

<template>
  <div class="page">
    <el-card class="main-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>Gingest Desktop</span>
          <el-tag type="success">Wails + Go + Vue3 + TS</el-tag>
        </div>
      </template>

      <div class="content">
        <el-alert
            title="第一阶段：先跑通 Go 与 Vue 通信"
            type="info"
            show-icon
            :closable="false"
        />

        <div class="actions">
          <el-button type="primary" :icon="Connection" @click="testGoCall">
            测试 Go 调用
          </el-button>

          <el-button type="warning" :icon="FolderOpened" @click="handleSelectDirectory">
            选择本地项目目录
          </el-button>
        </div>

        <el-descriptions title="运行结果" :column="1" border>
          <el-descriptions-item label="Go 返回">
            {{ message || '暂无' }}
          </el-descriptions-item>

          <el-descriptions-item label="选择目录">
            {{ selectedPath || '暂未选择' }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 24px;
  box-sizing: border-box;
}

.main-card {
  max-width: 960px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.actions {
  display: flex;
  gap: 12px;
}
</style>