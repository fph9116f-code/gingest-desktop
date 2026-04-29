<script setup lang="ts">
import { computed } from 'vue'
import {
  Document,
  Download,
} from '@element-plus/icons-vue'

const props = defineProps<{
  title: string
  content: string
}>()

const emit = defineEmits<{
  (e: 'copy'): void
  (e: 'save'): void
}>()

const MAX_DISPLAY_LENGTH = 100000

const previewContent = computed(() => {
  if (!props.content) return ''

  if (props.content.length <= MAX_DISPLAY_LENGTH) {
    return props.content
  }

  return (
      props.content.substring(0, MAX_DISPLAY_LENGTH) +
      '\n\n\n================================================\n' +
      '【内容过长，已截断预览】\n' +
      '此处只展示前 10 万字符，完整内容请点击「保存 XML」。\n' +
      '================================================'
  )
})
</script>

<template>
  <el-card shadow="never" class="preview-card">
    <template #header>
      <div class="card-header">
        <strong>XML 预览 - {{ title }}</strong>

        <div class="preview-actions">
          <el-button type="success" :icon="Document" @click="emit('copy')">
            复制 XML
          </el-button>

          <el-button type="warning" :icon="Download" @click="emit('save')">
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
</template>

<style scoped>
.preview-card {
  flex: 1;
  min-height: 180px;
  overflow: hidden;
}

.preview-card :deep(.el-card__body) {
  height: calc(100% - 58px);
  box-sizing: border-box;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.preview-actions {
  display: flex;
  align-items: center;
  gap: 8px;
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
</style>