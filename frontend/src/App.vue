<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  FolderOpened,
  Connection,
  Document,
  Download,
  Refresh,
} from '@element-plus/icons-vue'

import {
  Greet,
  SelectAndScanLocalDirectory,
} from '../wailsjs/go/main/App'

interface TreeNode {
  id?: number
  label: string
  isFile: boolean
  fullPath?: string
  content?: string
  children?: TreeNode[]
}

interface GingestResponse {
  projectName: string
  fileCount: number
  estimatedTokens: number
  formattedSize: string
  directoryTree: TreeNode[]
  content: string
  fullContent: string
}

const loading = ref(false)
const message = ref('')
const resultData = ref<GingestResponse | null>(null)
const filterText = ref('')

const MAX_DISPLAY_LENGTH = 100000

const previewContent = computed(() => {
  if (!resultData.value?.content) return ''

  if (resultData.value.content.length <= MAX_DISPLAY_LENGTH) {
    return resultData.value.content
  }

  return (
      resultData.value.content.substring(0, MAX_DISPLAY_LENGTH) +
      '\n\n\n================================================\n' +
      '【内容过长，已截断预览】\n' +
      '此处只展示前 10 万字符，完整内容后续会走文件导出。\n' +
      '================================================'
  )
})

const treeProps = {
  children: 'children',
  label: 'label',
}

const filterNode = (value: string, data: TreeNode) => {
  if (!value) return true
  return data.label.toLowerCase().includes(value.toLowerCase())
}

const testGoCall = async () => {
  try {
    message.value = await Greet('Gingest Desktop')
    ElMessage.success('Go 调用成功')
  } catch (error) {
    console.error(error)
    ElMessage.error('Go 调用失败')
  }
}

const handleScanLocal = async () => {
  loading.value = true
  resultData.value = null

  try {
    const response = await SelectAndScanLocalDirectory()

    if (!response || !response.projectName) {
      ElMessage.info('已取消选择目录')
      return
    }

    resultData.value = response as GingestResponse
    ElMessage.success(`扫描完成，共 ${response.fileCount} 个文件`)
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '扫描失败')
  } finally {
    loading.value = false
  }
}

const handleCopy = async () => {
  if (!resultData.value?.content) return

  try {
    await navigator.clipboard.writeText(resultData.value.content)
    ElMessage.success('已复制当前 XML 内容')
  } catch (error) {
    console.error(error)
    ElMessage.error('复制失败')
  }
}

const resetView = () => {
  resultData.value = null
  filterText.value = ''
}
</script>

<template>
  <div class="page">
    <el-container class="layout">
      <el-header class="header">
        <div class="title">
          <span>Gingest Desktop</span>
          <el-tag type="success">Wails + Go + Vue3 + TS</el-tag>
        </div>

        <div class="actions">
          <el-button :icon="Connection" @click="testGoCall">
            测试 Go
          </el-button>

          <el-button
              type="primary"
              :icon="FolderOpened"
              :loading="loading"
              @click="handleScanLocal"
          >
            选择并扫描本地项目
          </el-button>

          <el-button
              v-if="resultData"
              type="info"
              plain
              :icon="Refresh"
              @click="resetView"
          >
            重置
          </el-button>
        </div>
      </el-header>

      <el-main class="main" v-loading="loading" element-loading-text="正在扫描本地项目...">
        <template v-if="resultData">
          <div class="top-grid">
            <el-card shadow="never" class="summary-card">
              <template #header>
                <strong>项目摘要</strong>
              </template>

              <el-descriptions :column="1" border>
                <el-descriptions-item label="项目">
                  {{ resultData.projectName }}
                </el-descriptions-item>
                <el-descriptions-item label="文件数">
                  {{ resultData.fileCount }}
                </el-descriptions-item>
                <el-descriptions-item label="Tokens">
                  {{ resultData.estimatedTokens }}
                </el-descriptions-item>
                <el-descriptions-item label="大小">
                  {{ resultData.formattedSize }}
                </el-descriptions-item>
                <el-descriptions-item label="Go 返回">
                  {{ message || '暂无' }}
                </el-descriptions-item>
              </el-descriptions>
            </el-card>

            <el-card shadow="never" class="tree-card">
              <template #header>
                <div class="card-header">
                  <strong>目录结构</strong>
                  <el-input
                      v-model="filterText"
                      placeholder="搜索文件"
                      size="small"
                      clearable
                      class="search-input"
                  />
                </div>
              </template>

              <el-tree
                  ref="treeRef"
                  :data="resultData.directoryTree"
                  :props="treeProps"
                  :filter-node-method="filterNode"
                  node-key="id"
                  show-checkbox
                  check-on-click-node
                  class="file-tree"
              >
                <template #default="{ node, data }">
                  <span :class="data.isFile ? 'file-node' : 'folder-node'">
                    <el-icon>
                      <Document />
                    </el-icon>
                    <span>{{ node.label }}</span>
                  </span>
                </template>
              </el-tree>
            </el-card>
          </div>

          <el-card shadow="never" class="preview-card">
            <template #header>
              <div class="card-header">
                <strong>XML 预览</strong>

                <div>
                  <el-button type="success" :icon="Document" @click="handleCopy">
                    复制 XML
                  </el-button>

                  <el-button type="warning" :icon="Download" disabled>
                    下载 XML：下一步实现
                  </el-button>
                </div>
              </div>
            </template>

            <el-input
                type="textarea"
                readonly
                :model-value="previewContent"
                class="preview-textarea"
            />
          </el-card>
        </template>

        <el-card v-else shadow="never" class="empty-card">
          <div class="empty-content">
            <h2>Gingest Desktop</h2>
            <p>点击右上角「选择并扫描本地项目」，开始生成 AI 代码上下文。</p>
          </div>
        </el-card>
      </el-main>
    </el-container>
  </div>
</template>

<style scoped>
.page {
  height: 100vh;
  background: #f5f7fa;
}

.layout {
  height: 100%;
}

.header {
  height: 64px;
  background: #24292f;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
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

.main {
  height: calc(100vh - 64px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.top-grid {
  height: 280px;
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 16px;
  flex-shrink: 0;
}

.summary-card,
.tree-card,
.preview-card {
  overflow: hidden;
}

.tree-card :deep(.el-card__body) {
  height: calc(100% - 58px);
  overflow: auto;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.search-input {
  width: 220px;
}

.file-tree {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 13px;
}

.file-node,
.folder-node {
  display: flex;
  align-items: center;
  gap: 6px;
}

.folder-node {
  color: #409eff;
}

.file-node {
  color: #606266;
}

.preview-card {
  flex: 1;
  min-height: 0;
}

.preview-card :deep(.el-card__body) {
  height: calc(100% - 58px);
  box-sizing: border-box;
}

.preview-textarea {
  height: 100%;
}

.preview-textarea :deep(.el-textarea__inner) {
  height: 100%;
  resize: none;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  background: #fafafa;
}

.empty-card {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-content {
  text-align: center;
  color: #606266;
}
</style>