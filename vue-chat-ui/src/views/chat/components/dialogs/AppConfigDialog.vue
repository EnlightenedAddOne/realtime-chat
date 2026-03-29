<template>
  <el-dialog
    :model-value="modelValue"
    title="设置"
    width="420px"
    class="custom-dialog settings-dialog"
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="settings-container">
      <!-- 主题模式 Section -->
      <div class="settings-section">
        <div class="section-header">
          <el-icon><Sunny /></el-icon>
          <span class="section-title">主题模式</span>
        </div>
        <div class="theme-mode-grid">
          <div 
            v-for="mode in themeModes" 
            :key="mode.value"
            class="theme-mode-card"
            :class="{ active: themeMode === mode.value }"
            @click="handleModeChange(mode.value)"
          >
            <div class="mode-icon">
              <el-icon :size="24"><component :is="mode.icon" /></el-icon>
            </div>
            <span class="mode-name">{{ mode.label }}</span>
            <el-icon v-if="themeMode === mode.value" class="check-icon"><Check /></el-icon>
          </div>
        </div>
      </div>

      <!-- 主题配色 Section -->
      <div class="settings-section">
        <div class="section-header">
          <el-icon><Brush /></el-icon>
          <span class="section-title">主题配色</span>
        </div>
        <div class="color-theme-grid">
          <div 
            v-for="color in colorThemes" 
            :key="color.value"
            class="color-theme-card"
            :class="{ active: themeColor === color.value }"
            @click="handleColorChange(color.value)"
          >
            <div class="color-preview" :style="{ background: color.gradient }">
              <div class="color-accent" :style="{ background: color.accent }"></div>
            </div>
            <span class="color-name">{{ color.label }}</span>
            <el-icon v-if="themeColor === color.value" class="check-icon"><Check /></el-icon>
          </div>
        </div>
      </div>

      <!-- 实时预览 -->
      <div class="settings-preview">
        <div class="preview-label">预览效果</div>
        <div class="preview-card">
          <div class="preview-sidebar" :style="{ background: getPreviewGradient() }">
            <div class="preview-avatar"></div>
            <div class="preview-icons">
              <span></span><span></span><span></span>
            </div>
          </div>
          <div class="preview-content">
            <div class="preview-header"></div>
            <div class="preview-msgs">
              <div class="preview-msg left"></div>
              <div class="preview-msg right" :style="{ background: getPreviewGradient() }"></div>
            </div>
            <div class="preview-input"></div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button round @click="emit('update:modelValue', false)">关闭</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { Sunny, Moon, Monitor, Brush, Check } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true
  },
  themeMode: {
    type: String,
    required: true
  },
  themeColor: {
    type: String,
    required: true
  }
})

const emit = defineEmits([
  'update:modelValue',
  'update:themeMode',
  'update:themeColor',
  'theme-change',
  'theme-color-change'
])

const themeModes = [
  { value: 'light', label: '亮色模式', icon: Sunny },
  { value: 'dark', label: '暗色模式', icon: Moon },
  { value: 'system', label: '跟随系统', icon: Monitor }
]

const colorThemes = [
  { 
    value: 'default', 
    label: '粉蓝渐变', 
    gradient: 'linear-gradient(180deg, #a6c1ee 0%, #fbc2eb 100%)',
    accent: '#fbc2eb'
  },
  { 
    value: 'green', 
    label: '清新绿', 
    gradient: 'linear-gradient(180deg, #56ab2f 0%, #a8e063 100%)',
    accent: '#56ab2f'
  },
  { 
    value: 'purple', 
    label: '神秘紫', 
    gradient: 'linear-gradient(180deg, #6e48aa 0%, #9d50bb 100%)',
    accent: '#9d50bb'
  }
]

function handleModeChange(value) {
  emit('update:themeMode', value)
  emit('theme-change', value)
}

function handleColorChange(value) {
  emit('update:themeColor', value)
  emit('theme-color-change', value)
}

function getPreviewGradient() {
  const selected = colorThemes.find(c => c.value === props.themeColor)
  return selected?.gradient || colorThemes[0].gradient
}
</script>

<style scoped>
.settings-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 4px 0;
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--app-text-primary);
  font-weight: 600;
  font-size: 14px;
}

.section-header .el-icon {
  color: var(--app-color-secondary);
  font-size: 18px;
}

/* Theme Mode Grid */
.theme-mode-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.theme-mode-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 12px;
  background: var(--app-bg-secondary);
  border: 2px solid var(--app-border-light);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;
}

.theme-mode-card:hover {
  border-color: var(--app-color-secondary);
  transform: translateY(-2px);
  box-shadow: var(--app-shadow-sm);
}

.theme-mode-card.active {
  border-color: var(--app-color-secondary);
  background: var(--app-bg-active);
}

.mode-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--app-bg-tertiary);
  border-radius: 12px;
  color: var(--app-text-secondary);
  transition: all 0.25s ease;
}

.theme-mode-card.active .mode-icon {
  background: linear-gradient(135deg, var(--app-color-secondary) 0%, var(--app-color-primary) 100%);
  color: white;
}

.mode-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--app-text-primary);
}

.check-icon {
  position: absolute;
  top: 8px;
  right: 8px;
  color: var(--app-color-secondary);
  font-size: 14px;
  font-weight: bold;
}

/* Color Theme Grid */
.color-theme-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.color-theme-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: var(--app-bg-secondary);
  border: 2px solid var(--app-border-light);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;
}

.color-theme-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--app-shadow-sm);
}

.color-theme-card.active {
  border-color: var(--app-color-secondary);
}

.color-preview {
  width: 100%;
  height: 48px;
  border-radius: 8px;
  position: relative;
  overflow: hidden;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.1);
}

.color-accent {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 24px;
  height: 24px;
  border-radius: 6px 0 8px 0;
  box-shadow: -2px -2px 4px rgba(0,0,0,0.1);
}

.color-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--app-text-primary);
}

.color-theme-card .check-icon {
  top: 6px;
  right: 6px;
  color: white;
  text-shadow: 0 1px 2px rgba(0,0,0,0.3);
}

/* Preview Section */
.settings-preview {
  margin-top: 8px;
}

.preview-label {
  font-size: 12px;
  color: var(--app-text-secondary);
  margin-bottom: 10px;
}

.preview-card {
  display: flex;
  height: 100px;
  background: var(--app-bg-tertiary);
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--app-border-light);
}

.preview-sidebar {
  width: 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  gap: 6px;
}

.preview-avatar {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  background: rgba(255,255,255,0.4);
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.preview-icons {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 4px;
}

.preview-icons span {
  width: 14px;
  height: 14px;
  border-radius: 3px;
  background: rgba(255,255,255,0.3);
}

.preview-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 8px;
  gap: 6px;
  background: var(--app-bg-primary);
}

.preview-header {
  height: 12px;
  width: 60%;
  background: var(--app-bg-secondary);
  border-radius: 3px;
}

.preview-msgs {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.preview-msg {
  height: 16px;
  border-radius: 4px;
  width: 60%;
}

.preview-msg.left {
  background: var(--app-bg-secondary);
  width: 50%;
}

.preview-msg.right {
  align-self: flex-end;
  width: 55%;
}

.preview-input {
  height: 20px;
  background: var(--app-bg-secondary);
  border-radius: 4px;
}

/* Footer */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
}

.dialog-footer .el-button {
  min-width: 80px;
}
</style>
