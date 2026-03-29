<template>
  <aside class="mini-sidebar">
    <div class="mini-top">
      <div class="app-logo">
        <el-icon><ChatLineRound /></el-icon>
      </div>

      <el-avatar
        :size="40"
        :src="userInfo?.avatar_url"
        class="mini-avatar"
        shape="circle"
        title="个人信息"
        @click="emit('open-profile')"
      >
        {{ userInfo?.nickname?.[0] || 'U' }}
      </el-avatar>

      <div class="nav-group">
        <div class="nav-icon" :class="{ active: activeTab === 'messages' }" title="消息" @click="emit('update:activeTab', 'messages')">
          <el-icon><ChatDotRound /></el-icon>
        </div>
        <div class="nav-icon" :class="{ active: activeTab === 'contacts' }" title="联系人" @click="emit('update:activeTab', 'contacts')">
          <el-icon><User /></el-icon>
        </div>
        <div class="nav-icon" :class="{ active: activeTab === 'groups' }" title="群聊" @click="emit('update:activeTab', 'groups')">
          <el-icon><School /></el-icon>
        </div>
      </div>
    </div>

    <div class="mini-bottom">
      <el-popover
        placement="right-end"
        trigger="click"
        :width="180"
        :offset="12"
        :show-arrow="false"
        popper-class="settings-popover custom-popover-anim"
      >
        <template #reference>
          <div class="nav-icon settings-btn" title="更多">
            <el-icon><Menu /></el-icon>
          </div>
        </template>
        <div class="popover-menu">
          <div class="menu-item" @click="emit('open-app-config')">
            <el-icon><Setting /></el-icon> 设置
          </div>
          <div class="menu-divider"></div>
          <div class="menu-item danger" @click="emit('logout')">
            <el-icon><SwitchButton /></el-icon> 退出登录
          </div>
        </div>
      </el-popover>
    </div>
  </aside>
</template>

<script setup>
import { ChatDotRound, ChatLineRound, Menu, Setting, SwitchButton, User, School } from '@element-plus/icons-vue'

defineProps({
  userInfo: {
    type: Object,
    default: null
  },
  activeTab: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:activeTab', 'open-profile', 'open-app-config', 'logout'])
</script>
