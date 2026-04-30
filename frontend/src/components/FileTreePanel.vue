<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElTree } from 'element-plus'
import {
  Document,
  Folder,
  Operation,
  Refresh,
} from '@element-plus/icons-vue'
import type { SelectedFileStats, TreeNode } from '../types/gingest'
import { formatSize, formatToken, getTokenLevel } from '../utils/format'

const props = defineProps<{
  treeData: TreeNode[]
  selectedView: boolean
}>()

const emit = defineEmits<{
  (e: 'assemble-selected', files: TreeNode[]): void
  (e: 'restore-full'): void
  (e: 'selection-change', stats: SelectedFileStats): void
}>()

const treeRef = ref<InstanceType<typeof ElTree>>()
const filterText = ref('')
const selectedStats = ref<SelectedFileStats>({
  fileCount: 0,
  sizeBytes: 0,
  formattedSize: '0 B',
  estimatedTokens: 0,
})

const treeProps = {
  children: 'children',
  label: 'label',
}

const selectedSummaryText = computed(() => {
  return `${selectedStats.value.fileCount} 文件 / ${formatToken(selectedStats.value.estimatedTokens)} tokens / ${selectedStats.value.formattedSize}`
})

const filterNode = (value: string, data: any): boolean => {
  if (!value) return true

  const keyword = value.toLowerCase()
  const label = String(data?.label || '').toLowerCase()
  const fullPath = String(data?.fullPath || '').toLowerCase()

  return label.includes(keyword) || fullPath.includes(keyword)
}

watch(filterText, (value) => {
  treeRef.value?.filter(value)
})

watch(
    () => props.treeData,
    () => {
      clearChecked()
    },
)

const getCheckedFileNodes = () => {
  if (!treeRef.value) return []
  const checkedNodes = treeRef.value.getCheckedNodes(false, true) as TreeNode[]
  return checkedNodes.filter((node) => node.isFile)
}

const calculateSelectedStats = (files: TreeNode[]): SelectedFileStats => {
  const sizeBytes = files.reduce((sum, file) => sum + (file.sizeBytes || 0), 0)
  const estimatedTokens = files.reduce(
      (sum, file) => sum + (file.estimatedTokens || Math.floor((file.content?.length || 0) / 4)),
      0,
  )

  return {
    fileCount: files.length,
    sizeBytes,
    formattedSize: formatSize(sizeBytes),
    estimatedTokens,
  }
}

const updateSelectedStats = () => {
  const files = getCheckedFileNodes()
  const stats = calculateSelectedStats(files)
  selectedStats.value = stats
  emit('selection-change', stats)
}

const handleAssembleSelected = () => {
  const files = getCheckedFileNodes()
  emit('assemble-selected', files)
  updateSelectedStats()
}

const clearChecked = () => {
  treeRef.value?.setCheckedKeys([])
  selectedStats.value = {
    fileCount: 0,
    sizeBytes: 0,
    formattedSize: '0 B',
    estimatedTokens: 0,
  }
  emit('selection-change', selectedStats.value)
}
const getAllFileNodes = () => {
  const result: TreeNode[] = []

  const walk = (nodes: TreeNode[]) => {
    nodes.forEach((node) => {
      if (node.isFile) {
        result.push(node)
        return
      }

      if (node.children?.length) {
        walk(node.children)
      }
    })
  }

  walk(props.treeData || [])

  return result
}

const checkFilesByPaths = (paths: string[]) => {
  if (!treeRef.value) return

  const pathSet = new Set(
      paths
          .map((path) => path?.trim())
          .filter(Boolean),
  )

  const checkedKeys = getAllFileNodes()
      .filter((file) => pathSet.has(file.fullPath || file.label))
      .map((file) => file.id)
      .filter((id): id is number => typeof id === 'number')

  treeRef.value.setCheckedKeys(checkedKeys)
  updateSelectedStats()
}

const getFileTokenTagType = (node: TreeNode) => {
  const level = getTokenLevel(node.estimatedTokens || 0)

  if (level === 'danger') return 'danger'
  if (level === 'warning') return 'warning'
  return 'info'
}

defineExpose({
  clearChecked,
  getCheckedFileNodes,
  checkFilesByPaths,
})
</script>

<template>
  <el-card shadow="never" class="tree-card">
    <template #header>
      <div class="tree-header">
        <div class="tree-title-row">
          <strong>目录结构</strong>

          <el-tag
              :type="selectedStats.estimatedTokens >= 20000 ? 'danger' : selectedStats.estimatedTokens >= 8000 ? 'warning' : 'info'"
          >
            已勾选：{{ selectedSummaryText }}
          </el-tag>
        </div>

        <div class="tree-actions">
          <el-input
              v-model="filterText"
              placeholder="搜索文件 / 路径"
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
              v-if="selectedView"
              size="small"
              :icon="Refresh"
              @click="emit('restore-full')"
          >
            恢复全库
          </el-button>
        </div>
      </div>
    </template>

    <el-tree
        ref="treeRef"
        :data="treeData"
        :props="treeProps"
        :filter-node-method="filterNode"
        node-key="id"
        show-checkbox
        check-on-click-node
        class="file-tree"
        @check="updateSelectedStats"
        @check-change="updateSelectedStats"
    >
      <template #default="{ node, data }">
        <span :class="data.isFile ? 'file-node' : 'folder-node'">
          <span class="node-main">
            <el-icon>
              <Document v-if="data.isFile" />
              <Folder v-else />
            </el-icon>
            <span class="node-label">{{ node.label }}</span>
          </span>

          <span v-if="data.isFile" class="file-meta">
            <el-tag size="small" type="info" effect="plain">
              {{ data.formattedSize || '0 B' }}
            </el-tag>

            <el-tag
                size="small"
                :type="getFileTokenTagType(data)"
                effect="plain"
            >
              {{ formatToken(data.estimatedTokens || 0) }} tokens
            </el-tag>
          </span>
        </span>
      </template>
    </el-tree>
  </el-card>
</template>

<style scoped>
.tree-card {
  overflow: hidden;
}

.tree-card :deep(.el-card__body) {
  height: calc(100% - 74px);
  overflow: auto;
}

.tree-header {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tree-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.tree-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-input {
  width: 240px;
}

.file-tree {
  font-family: Consolas, "Courier New", monospace;
  font-size: 13px;
}

.file-node,
.folder-node {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.node-main {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.node-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-meta {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding-left: 10px;
}

.folder-node {
  color: #409eff;
}

.file-node {
  color: #606266;
}
</style>