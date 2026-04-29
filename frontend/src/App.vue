<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage, ElTree } from 'element-plus'
import {
  Connection,
  Document,
  Download,
  Folder,
  FolderOpened,
  Operation,
  Refresh,
  Setting,
} from '@element-plus/icons-vue'

import {
  GetFilterConfig,
  Greet,
  ResetFilterConfig,
  SaveFilterConfig,
  SaveXMLFile,
  SelectAndScanLocalDirectory,
} from '../wailsjs/go/main/App'
import { EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

interface TreeNode {
  id?: number
  label: string
  isFile: boolean
  fullPath?: string
  content?: string
  children?: TreeNode[]
}

interface SkipSample {
  reason: string
  path: string
}

interface ScanDiagnostics {
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
interface FilterConfig {
  ignoreDirectories: string[]
  ignoreExtensions: string[]
  ignoreFileNames: string[]
  maxFileCount: number
  maxTotalSizeMB: number
  maxSingleFileSizeMB: number
}
interface GingestResponse {
  projectName: string
  fileCount: number
  estimatedTokens: number
  formattedSize: string
  directoryTree: TreeNode[]
  content: string
  fullContent: string
  diagnostics?: ScanDiagnostics
}
interface ScanProgress {
  stage: string
  message: string
  currentPath: string
  processedFiles: number
  skippedFiles: number
  totalSize: number
  formattedSize: string
}



interface FilterConfigForm {
  ignoreDirectoriesText: string
  ignoreExtensionsText: string
  ignoreFileNamesText: string
  maxFileCount: number
  maxTotalSizeMB: number
  maxSingleFileSizeMB: number
}

const loading = ref(false)
const message = ref('')
const resultData = ref<GingestResponse | null>(null)
const fullResultData = ref<GingestResponse | null>(null)
const filterText = ref('')
const treeRef = ref<InstanceType<typeof ElTree>>()
const currentViewTitle = ref('全部提取结果')
const scanProgress = ref<ScanProgress | null>(null)

const settingsVisible = ref(false)
const settingsLoading = ref(false)
const filterConfig = ref<FilterConfig | null>(null)
const settingsForm = ref<FilterConfigForm>({
  ignoreDirectoriesText: '',
  ignoreExtensionsText: '',
  ignoreFileNamesText: '',
  maxFileCount: 3000,
  maxTotalSizeMB: 50,
  maxSingleFileSizeMB: 2,
})

const topHeight = ref(300)
const isResizing = ref(false)
let resizeStartY = 0
let resizeStartTopHeight = 300

const MAX_DISPLAY_LENGTH = 100000

const treeProps = {
  children: 'children',
  label: 'label',
}

const previewContent = computed(() => {
  if (!resultData.value?.content) return ''

  if (resultData.value.content.length <= MAX_DISPLAY_LENGTH) {
    return resultData.value.content
  }

  return (
      resultData.value.content.substring(0, MAX_DISPLAY_LENGTH) +
      '\n\n\n================================================\n' +
      '【内容过长，已截断预览】\n' +
      '此处只展示前 10 万字符，完整内容请点击「保存 XML」。\n' +
      '================================================'
  )
})

const skipReasonRows = computed(() => {
  const counts = resultData.value?.diagnostics?.skipReasonCounts || {}

  return Object.entries(counts)
      .map(([reason, count]) => ({
        reason,
        count,
      }))
      .sort((a, b) => b.count - a.count)
})

const scanProgressPercentage = computed(() => {
  if (!loading.value) return 100
  if (!scanProgress.value) return 5

  switch (scanProgress.value.stage) {
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
const hasAssembledView = computed(() => {
  return currentViewTitle.value !== '全部提取结果'
})

const currentSummary = computed(() => {
  if (!resultData.value) return null

  return {
    projectName: resultData.value.projectName,
    fileCount: resultData.value.fileCount,
    estimatedTokens: resultData.value.estimatedTokens,
    formattedSize: resultData.value.formattedSize,
  }
})

const normalizeLines = (text: string) => {
  return text
      .split('\n')
      .map((item) => item.trim())
      .filter(Boolean)
}

const configToForm = (config: FilterConfig): FilterConfigForm => {
  return {
    ignoreDirectoriesText: (config.ignoreDirectories || []).join('\n'),
    ignoreExtensionsText: (config.ignoreExtensions || []).join('\n'),
    ignoreFileNamesText: (config.ignoreFileNames || []).join('\n'),
    maxFileCount: Number(config.maxFileCount || 3000),
    maxTotalSizeMB: Number(config.maxTotalSizeMB || 50),
    maxSingleFileSizeMB: Number(config.maxSingleFileSizeMB || 2),
  }
}

const formToConfig = (): FilterConfig => {
  return {
    ignoreDirectories: normalizeLines(settingsForm.value.ignoreDirectoriesText),
    ignoreExtensions: normalizeLines(settingsForm.value.ignoreExtensionsText),
    ignoreFileNames: normalizeLines(settingsForm.value.ignoreFileNamesText),
    maxFileCount: Number(settingsForm.value.maxFileCount || 3000),
    maxTotalSizeMB: Number(settingsForm.value.maxTotalSizeMB || 50),
    maxSingleFileSizeMB: Number(settingsForm.value.maxSingleFileSizeMB || 2),
  }
}

const loadFilterConfig = async () => {
  settingsLoading.value = true

  try {
    const config = (await GetFilterConfig()) as FilterConfig
    filterConfig.value = config
    settingsForm.value = configToForm(config)
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '读取过滤配置失败')
  } finally {
    settingsLoading.value = false
  }
}

const openSettings = async () => {
  settingsVisible.value = true
  await loadFilterConfig()
}

const handleSaveFilterConfig = async () => {
  settingsLoading.value = true

  try {
    const saved = (await SaveFilterConfig(formToConfig())) as FilterConfig
    filterConfig.value = saved
    settingsForm.value = configToForm(saved)
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
    const config = (await ResetFilterConfig()) as FilterConfig
    filterConfig.value = config
    settingsForm.value = configToForm(config)
    ElMessage.success('已恢复默认过滤规则')
  } catch (error: any) {
    console.error(error)
    ElMessage.error(error?.message || '恢复默认配置失败')
  } finally {
    settingsLoading.value = false
  }
}

const filterNode = (value: string, data: any) => {
  if (!value) return true
  return data.label?.toLowerCase().includes(value.toLowerCase())
}

watch(filterText, (value) => {
  treeRef.value?.filter(value)
})

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
    const response = (await SelectAndScanLocalDirectory()) as GingestResponse

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

const escapeXmlAttribute = (value: string) => {
  return value
      .replaceAll('&', '&amp;')
      .replaceAll('"', '&quot;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
}

const safeCDATA = (value: string) => {
  return value.replaceAll(']]>', ']]]]><![CDATA[>')
}

const generateTreeText = (nodes: TreeNode[], prefix = ''): string => {
  let text = ''

  nodes.forEach((node, index) => {
    const isLast = index === nodes.length - 1
    const connector = isLast ? '└── ' : '├── '
    const childPrefix = prefix + (isLast ? '    ' : '│   ')

    text += prefix + connector + node.label + (node.isFile ? '' : '/') + '\n'

    if (node.children && node.children.length > 0) {
      text += generateTreeText(node.children, childPrefix)
    }
  })

  return text
}

const buildXmlByFiles = (files: TreeNode[], exportType: string) => {
  if (!fullResultData.value) return ''

  const estimatedTokens = Math.floor(
      files.reduce((sum, file) => sum + (file.content?.length || 0), 0) / 4,
  )

  let xml = ''
  xml += '<project_summary>\n'
  xml += `Project: ${fullResultData.value.projectName}\n`
  xml += `Export Type: ${exportType}\n`
  xml += `Total Files: ${files.length}\n`
  xml += `Estimated Tokens: ${estimatedTokens}\n`
  xml += '</project_summary>\n\n'

  xml += '<directory_tree>\n'
  xml += '.\n'
  xml += generateTreeText(fullResultData.value.directoryTree)
  xml += '</directory_tree>\n\n'

  xml += '<files>\n'

  files.forEach((file) => {
    xml += `<file path="${escapeXmlAttribute(file.fullPath || file.label)}">\n`
    xml += '<![CDATA[\n'
    xml += safeCDATA(file.content || '')
    xml += '\n]]>\n'
    xml += '</file>\n\n'
  })

  xml += '</files>'

  return xml
}

const handleAssembleSelected = () => {
  if (!fullResultData.value || !treeRef.value) return

  const checkedNodes = treeRef.value.getCheckedNodes(false, true) as TreeNode[]
  const selectedFiles = checkedNodes.filter((node) => node.isFile)

  if (selectedFiles.length === 0) {
    ElMessage.warning('请先在目录树中勾选需要组装的文件')
    return
  }

  const xml = buildXmlByFiles(selectedFiles, `Selected Files (${selectedFiles.length} files)`)

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
  ElMessage.success('已恢复全库 XML 视图')
}

const buildSuggestedFileName = () => {
  if (!resultData.value?.projectName) return 'gingest_export.xml'

  const suffix = hasAssembledView.value ? '_selected' : '_full'

  return (
      resultData.value.projectName
          .replace(/^Local:\s*/, '')
          .replace(/[\\/:*?"<>|]/g, '_') +
      `_gingest${suffix}.xml`
  )
}

const handleSaveXML = async () => {
  if (!resultData.value?.content) {
    ElMessage.warning('当前没有可保存的 XML 内容')
    return
  }

  try {
    const savedPath = await SaveXMLFile(resultData.value.content, buildSuggestedFileName())

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
  filterText.value = ''
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
      <el-header class="header">
        <div class="title">
          <span>Gingest Desktop</span>
          <el-tag type="success">Wails + Go + Vue3 + TS</el-tag>
        </div>

        <div class="actions">
          <el-button :icon="Connection" @click="testGoCall">
            测试 Go
          </el-button>

          <el-button :icon="Setting" @click="openSettings">
            过滤配置
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
              @click="clearResult"
          >
            清空
          </el-button>
        </div>
      </el-header>

      <el-main class="main">
        <el-card v-if="loading && scanProgress" shadow="never" class="progress-card">
          <div class="progress-header">
            <strong>扫描进度</strong>
            <el-tag type="info">
              {{ scanProgress.stage }}
            </el-tag>
          </div>

          <el-progress
              :percentage="scanProgressPercentage"
              :indeterminate="scanProgress.stage === 'scanning'"
          />

          <div class="progress-detail">
            <div>
              <strong>状态：</strong>
              {{ scanProgress.message || '准备中...' }}
            </div>

            <div>
              <strong>当前：</strong>
              <span class="current-path">{{ scanProgress.currentPath || '-' }}</span>
            </div>

            <div class="progress-stats">
              <span>已处理：{{ scanProgress.processedFiles || 0 }} 文件</span>
              <span>已跳过：{{ scanProgress.skippedFiles || 0 }} 项</span>
              <span>已读取：{{ scanProgress.formattedSize || '0 B' }}</span>
            </div>
          </div>
        </el-card>

        <template v-if="resultData">
          <div class="workbench">
            <div class="top-section" :style="{ height: topHeight + 'px' }">
              <div class="top-grid">
                <el-card shadow="never" class="summary-card">
                  <template #header>
                    <div class="card-header">
                      <strong>项目摘要</strong>
                      <el-tag v-if="hasAssembledView" type="warning">局部视图</el-tag>
                      <el-tag v-else type="success">全库视图</el-tag>
                    </div>
                  </template>

                  <el-descriptions v-if="currentSummary" :column="1" border>
                    <el-descriptions-item label="项目">
                      {{ currentSummary.projectName }}
                    </el-descriptions-item>
                    <el-descriptions-item label="文件数">
                      {{ currentSummary.fileCount }}
                    </el-descriptions-item>
                    <el-descriptions-item label="Tokens">
                      {{ currentSummary.estimatedTokens }}
                    </el-descriptions-item>
                    <el-descriptions-item label="大小">
                      {{ currentSummary.formattedSize }}
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
                      v-if="resultData.diagnostics?.stoppedEarly"
                      :title="resultData.diagnostics.stopReason"
                      type="warning"
                      show-icon
                      :closable="false"
                      class="diagnostics-alert"
                  />
                  <el-alert
                      v-if="resultData.fileCount === 0 && resultData.diagnostics?.noFileHint"
                      :title="resultData.diagnostics.noFileHint"
                      type="warning"
                      show-icon
                      :closable="false"
                      class="diagnostics-alert"
                  />

                  <div v-if="resultData.diagnostics" class="diagnostics-box">
                    <div class="diagnostics-title">扫描诊断</div>
                    <div class="diagnostics-grid">
                      <div>
                        <span>访问项</span>
                        <strong>{{ resultData.diagnostics.visitedItems }}</strong>
                      </div>
                      <div>
                        <span>有效文件</span>
                        <strong>{{ resultData.diagnostics.acceptedFiles }}</strong>
                      </div>
                      <div>
                        <span>跳过项</span>
                        <strong>{{ resultData.diagnostics.skippedItems }}</strong>
                      </div>
                    </div>
                  </div>
                </el-card>

                <el-card shadow="never" class="tree-card">
                  <template #header>
                    <div class="card-header">
                      <strong>目录结构</strong>

                      <div class="tree-actions">
                        <el-input
                            v-model="filterText"
                            placeholder="搜索文件"
                            size="small"
                            clearable
                            class="search-input"
                        />

                        <el-button
                            type="primary"
                            size="small"
                            :icon="Operation"
                            @click="handleAssembleSelected"
                        >
                          组装勾选
                        </el-button>

                        <el-button
                            v-if="hasAssembledView"
                            size="small"
                            :icon="Refresh"
                            @click="restoreFullView"
                        >
                          恢复全库
                        </el-button>
                      </div>
                    </div>
                  </template>

                  <el-tree
                      ref="treeRef"
                      :data="fullResultData?.directoryTree || resultData.directoryTree"
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
                          <Document v-if="data.isFile" />
                          <Folder v-else />
                        </el-icon>
                        <span>{{ node.label }}</span>
                      </span>
                    </template>
                  </el-tree>
                  <div v-if="resultData.fileCount === 0 && skipReasonRows.length > 0" class="skip-panel">
                    <div class="skip-title">跳过原因统计</div>

                    <el-table :data="skipReasonRows" size="small" height="180">
                      <el-table-column prop="reason" label="原因" min-width="180" />
                      <el-table-column prop="count" label="数量" width="80" align="right" />
                    </el-table>

                    <div
                        v-if="resultData.diagnostics?.skipSamples?.length"
                        class="skip-samples"
                    >
                      <div class="skip-title">跳过样例</div>
                      <div
                          v-for="item in resultData.diagnostics.skipSamples"
                          :key="item.reason + item.path"
                          class="skip-sample-item"
                      >
                        <el-tag size="small" type="info">{{ item.reason }}</el-tag>
                        <span>{{ item.path }}</span>
                      </div>
                    </div>
                  </div>
                </el-card>
              </div>
            </div>

            <div class="resize-divider" @mousedown="startResize">
              <div class="resize-line"></div>
              <span class="resize-label">拖动调整上下区域</span>
            </div>

            <el-card shadow="never" class="preview-card">
              <template #header>
                <div class="card-header">
                  <strong>XML 预览 - {{ currentViewTitle }}</strong>

                  <div class="preview-actions">
                    <el-button type="success" :icon="Document" @click="handleCopy">
                      复制 XML
                    </el-button>

                    <el-button type="warning" :icon="Download" @click="handleSaveXML">
                      保存 XML
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
          </div>
        </template>

        <el-card v-else shadow="never" class="empty-card">
          <div class="empty-content">
            <h2>Gingest Desktop</h2>
            <p>点击右上角「选择并扫描本地项目」，开始生成 AI 代码上下文。</p>
            <p class="empty-subtitle">建议先在「过滤配置」里确认目录、扩展名和文件大小限制。</p>
          </div>
        </el-card>
      </el-main>

      <el-drawer
          v-model="settingsVisible"
          title="过滤配置"
          size="520px"
          destroy-on-close
      >
        <div v-loading="settingsLoading" class="settings-panel">
          <el-alert
              title="配置会保存到本机用户配置目录，保存后下次扫描立即生效。"
              type="info"
              show-icon
              :closable="false"
          />

          <el-form label-position="top" class="settings-form">
            <el-row :gutter="12">
              <el-col :span="8">
                <el-form-item label="最大文件数">
                  <el-input-number
                      v-model="settingsForm.maxFileCount"
                      :min="1"
                      :max="100000"
                      controls-position="right"
                      class="full-width"
                  />
                </el-form-item>
              </el-col>

              <el-col :span="8">
                <el-form-item label="总大小上限 MB">
                  <el-input-number
                      v-model="settingsForm.maxTotalSizeMB"
                      :min="1"
                      :max="2048"
                      controls-position="right"
                      class="full-width"
                  />
                </el-form-item>
              </el-col>

              <el-col :span="8">
                <el-form-item label="单文件上限 MB">
                  <el-input-number
                      v-model="settingsForm.maxSingleFileSizeMB"
                      :min="1"
                      :max="512"
                      controls-position="right"
                      class="full-width"
                  />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="忽略目录">
              <el-input
                  v-model="settingsForm.ignoreDirectoriesText"
                  type="textarea"
                  :rows="8"
                  placeholder="每行一个目录，例如：node_modules"
              />
            </el-form-item>

            <el-form-item label="忽略扩展名">
              <el-input
                  v-model="settingsForm.ignoreExtensionsText"
                  type="textarea"
                  :rows="10"
                  placeholder="每行一个扩展名，例如：.png"
              />
            </el-form-item>

            <el-form-item label="忽略文件名">
              <el-input
                  v-model="settingsForm.ignoreFileNamesText"
                  type="textarea"
                  :rows="6"
                  placeholder="每行一个文件名，例如：package-lock.json"
              />
            </el-form-item>
          </el-form>

          <div class="settings-footer">
            <el-button @click="handleResetFilterConfig">
              恢复默认
            </el-button>
            <el-button @click="loadFilterConfig">
              重新读取
            </el-button>
            <el-button type="primary" @click="handleSaveFilterConfig">
              保存配置
            </el-button>
          </div>
        </div>
      </el-drawer>
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

.main {
  height: calc(100vh - 64px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
}

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

.summary-card,
.tree-card,
.preview-card {
  overflow: hidden;
}

.summary-card :deep(.el-card__body),
.tree-card :deep(.el-card__body) {
  height: calc(100% - 58px);
  overflow: auto;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.tree-actions,
.preview-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-input {
  width: 220px;
}

.file-tree {
  font-family: Consolas, "Courier New", monospace;
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

.preview-card {
  flex: 1;
  min-height: 180px;
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
  font-family: Consolas, "Courier New", monospace;
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

.settings-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.settings-form {
  margin-top: 16px;
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

.full-width {
  width: 100%;
}

.settings-footer {
  flex-shrink: 0;
  padding-top: 14px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  border-top: 1px solid #ebeef5;
}
</style>