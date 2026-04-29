<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { FilterConfig, FilterConfigForm } from '../types/gingest'

const props = defineProps<{
  modelValue: boolean
  loading: boolean
  config: FilterConfig | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'save', config: FilterConfig): void
  (e: 'reset'): void
  (e: 'reload'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const form = reactive<FilterConfigForm>({
  ignoreDirectoriesText: '',
  ignoreExtensionsText: '',
  ignoreFileNamesText: '',
  maxFileCount: 3000,
  maxTotalSizeMB: 50,
  maxSingleFileSizeMB: 2,
})

const normalizeLines = (text: string) => {
  return text
      .split('\n')
      .map((item) => item.trim())
      .filter(Boolean)
}

const syncConfigToForm = (config: FilterConfig | null) => {
  if (!config) return

  form.ignoreDirectoriesText = (config.ignoreDirectories || []).join('\n')
  form.ignoreExtensionsText = (config.ignoreExtensions || []).join('\n')
  form.ignoreFileNamesText = (config.ignoreFileNames || []).join('\n')
  form.maxFileCount = Number(config.maxFileCount || 3000)
  form.maxTotalSizeMB = Number(config.maxTotalSizeMB || 50)
  form.maxSingleFileSizeMB = Number(config.maxSingleFileSizeMB || 2)
}

const buildConfigFromForm = (): FilterConfig => {
  return {
    ignoreDirectories: normalizeLines(form.ignoreDirectoriesText),
    ignoreExtensions: normalizeLines(form.ignoreExtensionsText),
    ignoreFileNames: normalizeLines(form.ignoreFileNamesText),
    maxFileCount: Number(form.maxFileCount || 3000),
    maxTotalSizeMB: Number(form.maxTotalSizeMB || 50),
    maxSingleFileSizeMB: Number(form.maxSingleFileSizeMB || 2),
  }
}

watch(
    () => props.config,
    (config) => syncConfigToForm(config),
    { immediate: true, deep: true },
)

const handleSave = () => {
  emit('save', buildConfigFromForm())
}
</script>

<template>
  <el-drawer
      v-model="visible"
      title="过滤配置"
      size="520px"
      destroy-on-close
  >
    <div v-loading="loading" class="settings-panel">
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
                  v-model="form.maxFileCount"
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
                  v-model="form.maxTotalSizeMB"
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
                  v-model="form.maxSingleFileSizeMB"
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
              v-model="form.ignoreDirectoriesText"
              type="textarea"
              :rows="8"
              placeholder="每行一个目录，例如：node_modules"
          />
        </el-form-item>

        <el-form-item label="忽略扩展名">
          <el-input
              v-model="form.ignoreExtensionsText"
              type="textarea"
              :rows="10"
              placeholder="每行一个扩展名，例如：.png"
          />
        </el-form-item>

        <el-form-item label="忽略文件名">
          <el-input
              v-model="form.ignoreFileNamesText"
              type="textarea"
              :rows="6"
              placeholder="每行一个文件名，例如：package-lock.json"
          />
        </el-form-item>
      </el-form>

      <div class="settings-footer">
        <el-button @click="emit('reset')">
          恢复默认
        </el-button>
        <el-button @click="emit('reload')">
          重新读取
        </el-button>
        <el-button type="primary" @click="handleSave">
          保存配置
        </el-button>
      </div>
    </div>
  </el-drawer>
</template>

<style scoped>
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