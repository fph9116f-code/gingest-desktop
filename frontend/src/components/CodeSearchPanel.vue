<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Check,
  Close,
  Operation,
  Search,
} from '@element-plus/icons-vue'
import type { TreeNode } from '../types/gingest'
import { formatToken, getTokenLevel } from '../utils/format'

interface CodeSearchResult {
  file: TreeNode
  pathHit: boolean
  contentHitCount: number
  snippets: string[]
}

const props = defineProps<{
  treeData: TreeNode[]
}>()

const emit = defineEmits<{
  (e: 'check-results', files: TreeNode[]): void
  (e: 'assemble-results', files: TreeNode[]): void
}>()

const keyword = ref('')
const includePath = ref(true)
const includeContent = ref(true)

const MAX_RESULTS = 100
const MAX_SNIPPETS_PER_FILE = 1
const SNIPPET_RADIUS = 40

const flattenFiles = (nodes: TreeNode[]) => {
  const result: TreeNode[] = []

  const walk = (items: TreeNode[]) => {
    items.forEach((item) => {
      if (item.isFile) {
        result.push(item)
        return
      }

      if (item.children?.length) {
        walk(item.children)
      }
    })
  }

  walk(nodes || [])

  return result
}

const allFiles = computed(() => flattenFiles(props.treeData || []))
const normalizedKeyword = computed(() => keyword.value.trim().toLowerCase())

const countMatches = (source: string, target: string) => {
  if (!source || !target) return 0

  let count = 0
  let startIndex = 0

  while (true) {
    const index = source.indexOf(target, startIndex)
    if (index === -1) break

    count++
    startIndex = index + target.length
  }

  return count
}

const buildSnippets = (content: string, lowerContent: string, target: string) => {
  const snippets: string[] = []

  if (!content || !lowerContent || !target) {
    return snippets
  }

  let startIndex = 0

  while (snippets.length < MAX_SNIPPETS_PER_FILE) {
    const index = lowerContent.indexOf(target, startIndex)
    if (index === -1) break

    const from = Math.max(0, index - SNIPPET_RADIUS)
    const to = Math.min(content.length, index + target.length + SNIPPET_RADIUS)

    const prefix = from > 0 ? '...' : ''
    const suffix = to < content.length ? '...' : ''

    snippets.push(
        prefix +
        content
            .slice(from, to)
            .replace(/\s+/g, ' ')
            .trim() +
        suffix,
    )

    startIndex = index + target.length
  }

  return snippets
}

const searchResults = computed<CodeSearchResult[]>(() => {
  const target = normalizedKeyword.value

  if (!target) {
    return []
  }

  const results: CodeSearchResult[] = []

  allFiles.value.forEach((file) => {
    const pathText = `${file.fullPath || ''} ${file.label || ''}`.toLowerCase()
    const content = file.content || ''
    const lowerContent = content.toLowerCase()

    const pathHit = includePath.value && pathText.includes(target)
    const contentHitCount = includeContent.value
        ? countMatches(lowerContent, target)
        : 0

    if (!pathHit && contentHitCount === 0) {
      return
    }

    results.push({
      file,
      pathHit,
      contentHitCount,
      snippets: contentHitCount > 0
          ? buildSnippets(content, lowerContent, target)
          : [],
    })
  })

  return results
      .sort((a, b) => {
        const aScore = a.contentHitCount + (a.pathHit ? 3 : 0)
        const bScore = b.contentHitCount + (b.pathHit ? 3 : 0)

        if (bScore !== aScore) return bScore - aScore

        return (b.file.estimatedTokens || 0) - (a.file.estimatedTokens || 0)
      })
      .slice(0, MAX_RESULTS)
})

const resultFiles = computed(() => searchResults.value.map((item) => item.file))

const resultSummary = computed(() => {
  const files = resultFiles.value

  const totalTokens = files.reduce(
      (sum, file) => sum + (file.estimatedTokens || Math.floor((file.content?.length || 0) / 4)),
      0,
  )

  const totalHits = searchResults.value.reduce(
      (sum, item) => sum + item.contentHitCount + (item.pathHit ? 1 : 0),
      0,
  )

  return {
    fileCount: files.length,
    totalTokens,
    totalHits,
  }
})

const getTokenTagType = (tokens: number) => {
  const level = getTokenLevel(tokens)

  if (level === 'danger') return 'danger'
  if (level === 'warning') return 'warning'

  return 'info'
}

const handleCheckResults = () => {
  if (resultFiles.value.length === 0) return
  emit('check-results', resultFiles.value)
}

const handleAssembleResults = () => {
  if (resultFiles.value.length === 0) return
  emit('assemble-results', resultFiles.value)
}

const clearSearch = () => {
  keyword.value = ''
}
</script>

<template>
  <div class="code-search-panel">
    <div class="search-top">
      <div class="search-title">
        <el-icon>
          <Search />
        </el-icon>
        <strong>代码搜索</strong>

        <el-tag v-if="normalizedKeyword" size="small" type="info">
          {{ resultSummary.fileCount }} 文件 / {{ resultSummary.totalHits }} 命中 / {{ formatToken(resultSummary.totalTokens) }} tokens
        </el-tag>
      </div>

      <div class="search-actions">
        <el-button
            size="small"
            type="primary"
            plain
            :icon="Check"
            :disabled="resultFiles.length === 0"
            @click="handleCheckResults"
        >
          勾选
        </el-button>

        <el-button
            size="small"
            type="success"
            plain
            :icon="Operation"
            :disabled="resultFiles.length === 0"
            @click="handleAssembleResults"
        >
          组装
        </el-button>

        <el-button
            size="small"
            text
            :icon="Close"
            :disabled="!keyword"
            @click="clearSearch"
        >
          清空
        </el-button>
      </div>
    </div>

    <div class="search-line">
      <el-input
          v-model="keyword"
          clearable
          size="small"
          placeholder="搜索类名、方法名、接口名、错误信息、业务关键词..."
          class="search-input"
      >
        <template #prefix>
          <el-icon>
            <Search />
          </el-icon>
        </template>
      </el-input>

      <el-checkbox v-model="includePath" size="small">
        路径
      </el-checkbox>

      <el-checkbox v-model="includeContent" size="small">
        内容
      </el-checkbox>
    </div>

    <div v-if="normalizedKeyword && searchResults.length > 0" class="result-list">
      <div
          v-for="item in searchResults"
          :key="item.file.fullPath || item.file.label"
          class="result-item"
      >
        <div class="result-row">
          <div class="result-path">
            {{ item.file.fullPath || item.file.label }}
          </div>

          <div class="result-tags">
            <el-tag v-if="item.pathHit" size="small" type="success" effect="plain">
              路径
            </el-tag>

            <el-tag
                v-if="item.contentHitCount > 0"
                size="small"
                type="warning"
                effect="plain"
            >
              {{ item.contentHitCount }} 次
            </el-tag>

            <el-tag
                size="small"
                :type="getTokenTagType(item.file.estimatedTokens || 0)"
                effect="plain"
            >
              {{ formatToken(item.file.estimatedTokens || 0) }}
            </el-tag>
          </div>
        </div>

        <div v-if="item.snippets.length > 0" class="snippet">
          {{ item.snippets[0] }}
        </div>
      </div>
    </div>

    <div v-else-if="normalizedKeyword && searchResults.length === 0" class="empty-result">
      没有匹配到文件
    </div>
  </div>
</template>

<style scoped>
.code-search-panel {
  flex-shrink: 0;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fff;
  padding: 8px 10px;
}

.search-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.search-title {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #303133;
}

.search-title strong {
  flex-shrink: 0;
}

.search-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.search-line {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-input {
  flex: 1;
  min-width: 220px;
}

.result-list {
  margin-top: 8px;
  max-height: 120px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-right: 4px;
}

.result-item {
  padding: 6px 8px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fafafa;
}

.result-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.result-path {
  min-width: 0;
  flex: 1;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-tags {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 4px;
}

.snippet {
  margin-top: 4px;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-result {
  margin-top: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  background: #fafafa;
  color: #909399;
  font-size: 12px;
}
</style>