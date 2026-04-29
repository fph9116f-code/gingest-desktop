<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElTree } from 'element-plus'
import {
  Document,
  Folder,
  Operation,
  Refresh,
} from '@element-plus/icons-vue'
import type { TreeNode } from '../types/gingest'

const props = defineProps<{
  treeData: TreeNode[]
  selectedView: boolean
}>()

const emit = defineEmits<{
  (e: 'assemble-selected', files: TreeNode[]): void
  (e: 'restore-full'): void
}>()

const treeRef = ref<InstanceType<typeof ElTree>>()
const filterText = ref('')

const treeProps = {
  children: 'children',
  label: 'label',
}

const filterNode = (value: string, data: any) => {
  if (!value) return true
  return data.label?.toLowerCase().includes(value.toLowerCase())
}

watch(filterText, (value) => {
  treeRef.value?.filter(value)
})

const getCheckedFileNodes = () => {
  if (!treeRef.value) return []
  const checkedNodes = treeRef.value.getCheckedNodes(false, true) as TreeNode[]
  return checkedNodes.filter((node) => node.isFile)
}

const handleAssembleSelected = () => {
  emit('assemble-selected', getCheckedFileNodes())
}

const clearChecked = () => {
  treeRef.value?.setCheckedKeys([])
}

defineExpose({
  clearChecked,
  getCheckedFileNodes,
})
</script>

<template>
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
  </el-card>
</template>

<style scoped>
.tree-card {
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
  gap: 12px;
}

.tree-actions {
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
</style>