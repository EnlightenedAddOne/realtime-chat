<template>
  <el-dialog
    :model-value="modelValue"
    title="个人信息设置"
    width="380px"
    class="custom-dialog profile-settings-dialog"
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="profile-settings-container">
      <div class="profile-hero">
        <el-avatar
          :size="80"
          :src="settingsForm.avatar_url || userInfo?.avatar_url"
          shape="circle"
          class="shadow-avatar"
        >
          {{ settingsForm.nickname?.[0] || userInfo?.username?.[0] }}
        </el-avatar>
        <el-button
          class="profile-avatar-edit-btn"
          circle
          size="small"
          :loading="updatingProfileAvatar"
          title="修改头像"
          @click="emit('select-avatar')"
        >
          <el-icon><Edit /></el-icon>
        </el-button>
      </div>

      <el-form :model="settingsForm" label-position="top">
        <div class="dialog-card-item">
          <el-form-item label="用户昵称" style="margin-bottom: 0;">
            <el-input
              v-model="settingsForm.nickname"
              placeholder="请输入您的昵称"
              maxlength="30"
              class="anime-input"
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>
        </div>
        <div class="settings-helper-text">
          * 支持上传 jpg/png/webp 格式图片作为头像
        </div>
      </el-form>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button round @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" round :disabled="updatingProfileAvatar" @click="emit('save')">保存</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { Edit, User } from '@element-plus/icons-vue'

defineProps({
  modelValue: {
    type: Boolean,
    required: true
  },
  settingsForm: {
    type: Object,
    required: true
  },
  userInfo: {
    type: Object,
    default: null
  },
  updatingProfileAvatar: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'select-avatar', 'save'])
</script>
