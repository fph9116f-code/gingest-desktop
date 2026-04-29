<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

import AppHeader from './components/AppHeader.vue'
import FileTreePanel from './components/FileTreePanel.vue'
import ScanProgressCard from './components/ScanProgressCard.vue'
import SettingsDrawer from './components/SettingsDrawer.vue'
import SummaryPanel from './components/SummaryPanel.vue'
import XmlPreviewPanel from './components/XmlPreviewPanel.vue'

import { gingestApi } from './api/gingestApi'
import { buildSuggestedXmlFileName, buildXmlByFiles } from './utils/xml'
import type { FilterConfig, GingestResponse, ScanProgress, TreeNode } from './types/gingest'
import { EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

const loading = ref(false)
const settingsVisible = ref(false)
const settingsLoading = ref(false)

const message = ref('')
const resultData = ref<GingestResponse | null>(null)
const fullResultData = ref<GingestResponse | null>(null)
const filterConfig = ref<FilterConfig | null>(null)
const scanProgress = ref<ScanProgress | null>(null)

const currentViewTitle = ref('全部提取结果')
const topHeight = ref(300)
const isResizing = ref(false)
const fileTreePanelRef = ref<InstanceType<typeof FileTreePanel>>()

let resizeStartY = 0
let resizeStartTopHeight = 300

const hasResult = computed(() => !!resultData.value)
const hasSelectedView = computed(() => currentViewTitle.value !== '全部提取结果')

const treeData = computed(() => {
  return fullResultData.value?.directoryTree || resultData.value?.directoryTree || []
})

const openSettings = async () => {
  settingsVisible.value = true
  await loadFilterConfig()
}

const loadFilterConfig = async () => {
  settingsLoading.value = true

  try {
    filterConfig.value = await gingestApi.getFilterConfig()
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '读取过滤配置失败')
  } finally {
    settingsLoading.value = false
  }
}

const handleSaveFilterConfig = async (config: FilterConfig) => {
  settingsLoading.value = true

  try {
    filterConfig.value = await gingestApi.saveFilterConfig(config)
    ElMessage.success('过滤规则已保存，下次扫描会自动生效')
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '保存过滤配置失败')
  } finally {
    settingsLoading.value = false
  }
}

const handleResetFilterConfig = async () => {
  settingsLoading.value = true

  try {
    filterConfig.value = await gingestApi.resetFilterConfig()
    ElMessage.success('已恢复默认过滤规则')
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '恢复默认配置失败')
  } finally {
    settingsLoading.value = false
  }
}

const handleTestGo = async () => {
  try {
    message.value = await gingestApi.greet('Gingest Desktop')
    ElMessage.success('Go 调用成功')
  } catch (error) {
    console.error(error)
    ElMessage.error('Go 调用失败')
  }
}

const handleScanLocal = async () => {
  loading.value = true
  resultData.value = null
  fullResultData.value = null
  currentViewTitle.value = '全部提取结果'

  scanProgress.value = {
    stage: 'start',
    message: '准备选择目录',
    currentPath: '',
    processedFiles: 0,
    skippedFiles: 0,
    totalSize: 0,
    formattedSize: '0 B',
  }

  try {
    const response = await gingestApi.scanLocalProject()

    if (!response || !response.projectName) {
      ElMessage.info('已取消选择目录')
      scanProgress.value = null
      return
    }

    resultData.value = response
    fullResultData.value = JSON.parse(JSON.stringify(response))
    currentViewTitle.value = '全部提取结果'

    if (response.diagnostics?.stoppedEarly) {
      ElMessage.warning(response.diagnostics.stopReason || '扫描已因配置限制提前停止')
    } else if (response.fileCount === 0) {
      ElMessage.warning(response.diagnostics?.noFileHint || '扫描完成，但没有匹配到有效文件')
    } else {
      ElMessage.success(`扫描完成，共 ${response.fileCount} 个文件`)
    }
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '扫描失败')
  } finally {
    loading.value = false
    scanProgress.value = null
  }
}

const handleAssembleSelected = (selectedFiles: TreeNode[]) => {
  if (!fullResultData.value) return

  if (selectedFiles.length === 0) {
    ElMessage.warning('请先在目录树中勾选需要组装的文件')
    return
  }

  const xml = buildXmlByFiles(
      fullResultData.value.projectName,
      fullResultData.value.directoryTree,
      selectedFiles,
      `Selected Files (${selectedFiles.length} files)`,
  )

  resultData.value = {
    ...fullResultData.value,
    fileCount: selectedFiles.length,
    estimatedTokens: Math.floor(xml.length / 4),
    formattedSize: `${(new Blob([xml]).size / 1024).toFixed(2)} KB`,
    content: xml,
  }

  currentViewTitle.value = `已组装勾选文件：${selectedFiles.length} 个`
  ElMessage.success(`组装完成，共 ${selectedFiles.length} 个文件`)
}

const restoreFullView = () => {
  if (!fullResultData.value) return

  resultData.value = JSON.parse(JSON.stringify(fullResultData.value))
  currentViewTitle.value = '全部提取结果'
  fileTreePanelRef.value?.clearChecked()
  ElMessage.success('已恢复全库 XML 视图')
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

const handleSaveXML = async () => {
  if (!resultData.value?.content) {
    ElMessage.warning('当前没有可保存的 XML 内容')
    return
  }

  try {
    const savedPath = await gingestApi.saveXmlFile(
        resultData.value.content,
        buildSuggestedXmlFileName(resultData.value.projectName, hasSelectedView.value),
    )

    if (!savedPath) {
      ElMessage.info('已取消保存')
      return
    }

    ElMessage.success(`保存成功：${savedPath}`)
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '保存失败')
  }
}

const clearResult = () => {
  resultData.value = null
  fullResultData.value = null
  currentViewTitle.value = '全部提取结果'
  scanProgress.value = null
}

const startResize = (event: MouseEvent) => {
  if (!resultData.value) return

  isResizing.value = true
  resizeStartY = event.clientY
  resizeStartTopHeight = topHeight.value

  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.userSelect = 'none'
}

const handleResize = (event: MouseEvent) => {
  if (!isResizing.value) return

  const deltaY = event.clientY - resizeStartY
  const maxHeight = Math.max(220, window.innerHeight - 360)
  const nextHeight = Math.min(Math.max(resizeStartTopHeight + deltaY, 180), maxHeight)

  topHeight.value = nextHeight
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.userSelect = ''
}

onMounted(async () => {
  EventsOn('scan-progress', (progress: ScanProgress) => {
    scanProgress.value = progress
  })

  await loadFilterConfig()
})

onUnmounted(() => {
  EventsOff('scan-progress')
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
})
</script>

<template>
  <div class="page">
    <el-container class="layout">
      <AppHeader
          :loading="loading"
          :has-result="hasResult"
          @test-go="handleTestGo"
          @open-settings="openSettings"
          @scan-local="handleScanLocal"
          @clear-result="clearResult"
      />

      <el-main class="main">
        <ScanProgressCard
            :loading="loading"
            :progress="scanProgress"
        />

        <template v-if="resultData">
          <div class="workbench">
            <div class="top-section" :style="{ height: topHeight + 'px' }">
              <div class="top-grid">
                <SummaryPanel
                    :result="resultData"
                    :filter-config="filterConfig"
                    :message="message"
                    :is-selected-view="hasSelectedView"
                />

                <FileTreePanel
                    ref="fileTreePanelRef"
                    :tree-data="treeData"
                    :selected-view="hasSelectedView"
                    @assemble-selected="handleAssembleSelected"
                    @restore-full="restoreFullView"
                />
              </div>
            </div>

            <div class="resize-divider" @mousedown="startResize">
              <div class="resize-line"></div>
              <span class="resize-label">拖动调整上下区域</span>
            </div>

            <XmlPreviewPanel
                :title="currentViewTitle"
                :content="resultData.content"
                @copy="handleCopy"
                @save="handleSaveXML"
            />
          </div>
        </template>

        <el-card v-else shadow="never" class="empty-card">
          <div class="empty-content">
            <h2>Gingest Desktop</h2>
            <p>点击右上角「选择并扫描本地项目」，开始生成 AI 代码上下文。</p>
            <p class="empty-subtitle">Facade 入口树已排除，当前只保留通用代码上下文构建能力。</p>
          </div>
        </el-card>
      </el-main>

      <SettingsDrawer
          v-model="settingsVisible"
          :loading="settingsLoading"
          :config="filterConfig"
          @save="handleSaveFilterConfig"
          @reset="handleResetFilterConfig"
          @reload="loadFilterConfig"
      />
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
  display: flex;
  flex-direction: column;
}

.main {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
}

.workbench {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.top-section {
  flex-shrink: 0;
  min-height: 180px;
}

.top-grid {
  height: 100%;
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 16px;
}

.resize-divider {
  height: 18px;
  flex-shrink: 0;
  cursor: row-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  user-select: none;
}

.resize-line {
  width: 80px;
  height: 4px;
  border-radius: 999px;
  background: #dcdfe6;
  transition: background 0.2s;
}

.resize-label {
  font-size: 12px;
  color: #909399;
}

.resize-divider:hover .resize-line {
  background: #409eff;
}

.empty-card {
  flex: 1;
  width: 100%;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-card :deep(.el-card__body) {
  width: 100%;
}

.empty-content {
  text-align: center;
  color: #606266;
}

.empty-subtitle {
  color: #909399;
  font-size: 13px;
}
</style>