<script setup lang="ts">
import { FolderOpened, UploadFilled } from '@element-plus/icons-vue'

defineProps<{
  active: boolean
  loading: boolean
}>()
</script>

<template>
  <transition name="drop-fade">
    <div v-if="active" class="drop-overlay">
      <div class="drop-card">
        <el-icon class="drop-icon">
          <UploadFilled v-if="!loading" />
          <FolderOpened v-else />
        </el-icon>

        <div class="drop-title">
          {{ loading ? '正在扫描中...' : '松开鼠标开始扫描项目目录' }}
        </div>

        <div class="drop-desc">
          {{ loading ? '当前已有扫描任务，请等待完成。' : '请拖入一个本地项目文件夹，Gingest 会自动生成 AI 代码上下文。' }}
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.drop-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(36, 41, 47, 0.72);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.drop-card {
  width: 460px;
  min-height: 220px;
  border: 2px dashed #79bbff;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.22);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: #303133;
}

.drop-icon {
  font-size: 56px;
  color: #409eff;
}

.drop-title {
  font-size: 22px;
  font-weight: 800;
}

.drop-desc {
  width: 340px;
  text-align: center;
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
}

.drop-fade-enter-active,
.drop-fade-leave-active {
  transition: opacity 0.16s ease;
}

.drop-fade-enter-from,
.drop-fade-leave-to {
  opacity: 0;
}
</style>