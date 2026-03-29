<template>
  <aside class="list-sidebar" :style="{ width: `${sidebarWidth}px` }">
    <div class="list-header">
      <div class="search-row">
        <el-input
          v-model="searchProxy"
          placeholder="搜索"
          prefix-icon="Search"
          class="anime-search"
          clearable
        />
        <el-tooltip content="添加好友" placement="bottom" :show-after="500">
          <div class="add-friend-btn" @click="emit('open-add-friend')">
            <el-icon><Plus /></el-icon>
          </div>
        </el-tooltip>
        <el-tooltip content="创建群组" placement="bottom" :show-after="500">
          <div class="add-friend-btn" @click="emit('open-create-group')">
            <el-icon><ChatLineRound /></el-icon>
          </div>
        </el-tooltip>
      </div>
    </div>

    <div class="friend-list custom-scrollbar">
      <div
        v-for="item in sidebarList"
        :key="item.isGroup ? `g${item.id}` : `u${item.id}`"
        class="friend-item"
        :class="{ active: currentUser?.id === item.id && currentUser?.isGroup === item.isGroup }"
        @click="emit('select-chat', item)"
      >
        <div class="friend-item-content">
          <div class="avatar-wrapper">
            <el-avatar :size="40" :src="item.avatar" class="friend-avatar" :shape="item.isGroup ? 'square' : 'circle'">
              {{ item.name?.[0] || 'U' }}
            </el-avatar>
            <div v-if="item.unread > 0" class="unread-badge">{{ item.unread > 99 ? '99+' : item.unread }}</div>
          </div>

          <div class="friend-details">
            <div class="friend-top-row" :style="activeTab !== 'messages' ? { marginBottom: 0, height: '100%', alignItems: 'center' } : {}">
              <span class="friend-name">{{ item.name }}</span>
              <span v-if="activeTab === 'messages'" class="friend-time">{{ formatTime(item.time) }}</span>
            </div>
            <div v-if="activeTab === 'messages'" class="friend-message-preview">{{ item.lastMsg || '暂无消息' }}</div>
            <div v-else-if="activeTab === 'contacts'" class="friend-message-preview">{{ item.target?.signature || '无个签' }}</div>
            <div v-else class="friend-message-preview">{{ item.target?.announcement || '暂无群公告' }}</div>
          </div>

          <el-tooltip v-if="activeTab === 'contacts'" content="删除好友" placement="top">
            <el-button
              circle
              text
              type="danger"
              class="contact-delete-btn"
              :loading="!!deletingFriendIds[item.id]"
              @click.stop="emit('delete-friend', item)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
      </div>

      <div v-if="sidebarList.length === 0" class="list-empty">
        <span>{{ activeTab === 'messages' ? '暂无消息' : (activeTab === 'contacts' ? '暂无联系人' : '暂无群聊') }}</span>
      </div>
    </div>

    <div class="resizer" @mousedown="emit('start-resize', $event)"></div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { ChatLineRound, Delete, Plus } from '@element-plus/icons-vue'

const props = defineProps({
  sidebarWidth: {
    type: Number,
    required: true
  },
  searchText: {
    type: String,
    required: true
  },
  sidebarList: {
    type: Array,
    required: true
  },
  currentUser: {
    type: Object,
    default: null
  },
  activeTab: {
    type: String,
    required: true
  },
  deletingFriendIds: {
    type: Object,
    default: () => ({})
  },
  formatTime: {
    type: Function,
    required: true
  }
})

const emit = defineEmits([
  'update:searchText',
  'open-add-friend',
  'open-create-group',
  'select-chat',
  'delete-friend',
  'start-resize'
])

const searchProxy = computed({
  get: () => props.searchText,
  set: (val) => emit('update:searchText', val)
})
</script>
