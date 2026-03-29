<template>
  <main class="chat-main">
    <template v-if="currentUser">
      <header class="chat-header">
        <div class="header-info">
          <h2 class="chat-title">
            {{ currentUser.nickname || currentUser.username || currentUser.name }}
            <span v-if="currentUser.isGroup" style="font-size: 14px; color: #9ca3af; font-weight: normal;">(群组)</span>
          </h2>
          <div v-if="!currentUser.isGroup" class="chat-status-text">
            <span class="status-dot" :class="{ online: currentUserOnline, offline: !currentUserOnline }"></span>
            <span v-if="currentUserOnlineLoading">状态获取中...</span>
            <span v-else>{{ currentUserOnline ? '在线' : '离线' }}</span>
          </div>
        </div>
        <div class="header-actions">
          <el-button
            v-if="!currentUser.isGroup"
            circle
            icon="VideoCamera"
            text
            class="header-more-btn"
            title="视频通话"
            @click="emit('start-call')"
          ></el-button>
          <el-button circle icon="More" text class="header-more-btn" @click="emit('more-click')"></el-button>
        </div>
      </header>

      <div class="messages-area custom-scrollbar" :ref="setMsgListRef">
        <!-- Theme backgrounds -->
        <div class="chat-bg-decoration"></div>
        
        <div
          v-for="(msg, index) in currentMessages"
          :key="index"
          class="message-wrapper"
          :class="msg.sender_id == userInfo?.id ? 'is-self' : 'other'"
        >
          <el-avatar
            v-if="msg.sender_id != userInfo?.id"
            :size="36"
            :src="msg.sender?.avatar_url || (currentUser.isGroup ? '' : currentUser.avatar_url)"
            class="message-avatar other"
          >
            {{ msg.sender?.nickname?.[0] || currentUser.nickname?.[0] || 'U' }}
          </el-avatar>

          <el-avatar
            v-else
            :size="36"
            :src="userInfo?.avatar_url"
            class="message-avatar self"
          >
            {{ userInfo?.nickname?.[0] || 'U' }}
          </el-avatar>

          <div class="message-content">
            <span v-if="currentUser.isGroup && msg.sender_id != userInfo?.id" class="sender-name">
              {{ msg.sender?.nickname || msg.sender?.username || 'Unknown' }}
            </span>

            <div v-if="msg.msg_type === 1" class="message-bubble text">
              {{ msg.content }}
            </div>
            <div v-else-if="msg.msg_type === 2" class="message-bubble image">
              <img :src="msg.content" loading="lazy" @click="emit('preview-image', msg.content)" />
            </div>
            <div v-else-if="msg.msg_type === 3" class="message-bubble voice" :class="{ 'playing': playingUrl === msg.content }" @click="emit('play-voice', msg.content)">
              <div class="voice-icon">
                <el-icon v-if="playingUrl === msg.content"><VideoPause /></el-icon>
                <el-icon v-else><Microphone /></el-icon>
              </div>
              <span class="voice-duration">{{ msg.duration || 0 }}s</span>
              <div v-if="playingUrl === msg.content" class="voice-wave-anim">
                <span class="bar"></span>
                <span class="bar"></span>
                <span class="bar"></span>
              </div>
            </div>
            <div v-else-if="msg.msg_type === 4" class="message-bubble video">
              <video :src="msg.content" controls />
            </div>
          </div>
        </div>
      </div>

      <footer class="input-area">
        <div class="toolbar">
          <el-popover placement="top-start" :width="340" trigger="click" popper-class="emoji-popover">
            <template #reference>
              <div class="tool-icon" title="表情"><el-icon><ChatDotRound /></el-icon></div>
            </template>
            <div class="emoji-grid custom-scrollbar">
              <span v-for="emoji in emojis" :key="emoji" class="emoji-item" @click="emit('emoji-click', emoji)">{{ emoji }}</span>
            </div>
          </el-popover>

          <el-tooltip content="图片" placement="top"><div class="tool-icon" @click="emit('select-image')"><el-icon><Picture /></el-icon></div></el-tooltip>
          <el-tooltip content="视频" placement="top"><div class="tool-icon" @click="emit('select-video')"><el-icon><VideoCamera /></el-icon></div></el-tooltip>
          <el-tooltip content="语音消息" placement="top"><div class="tool-icon" :class="{ 'is-recording': isRecording }" @click="emit('toggle-voice')"><el-icon><Microphone /></el-icon></div></el-tooltip>
        </div>

        <div v-if="!isRecording" class="input-row">
          <textarea
            v-model="inputProxy"
            placeholder="输入消息..."
            class="anime-input-textarea custom-scrollbar"
            @keydown.enter.exact.prevent="emit('send-text')"
          ></textarea>
          <button class="anime-send-btn" :disabled="!inputText.trim()" @click="emit('send-text')">
            <el-icon><Promotion /></el-icon>
          </button>
        </div>

        <div v-else class="recording-overlay">
          <div class="recording-info">
            <div class="recording-dot"></div>
            <span class="recording-timer">{{ formattedDuration }}</span>
            <span class="recording-status">正在录音...</span>
          </div>
          <div class="recording-controls">
            <el-tooltip content="取消" placement="top">
              <div class="record-btn cancel" @click="emit('cancel-recording')">
                <el-icon><Close /></el-icon>
              </div>
            </el-tooltip>
            <el-tooltip content="发送" placement="top">
              <div class="record-btn send" @click="emit('send-voice')">
                <el-icon><Check /></el-icon>
              </div>
            </el-tooltip>
          </div>
        </div>
      </footer>
    </template>

    <div v-else class="empty-state">
      <!-- Theme backgrounds -->
      <div class="chat-bg-decoration"></div>

      <video :src="welcomeImg" alt="Welcome" class="welcome-image" autoplay muted />
      <h3>欢迎使用 Chat</h3>
      <p>选择好友开始聊天</p>
    </div>
  </main>
</template>

<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import defaultThemeImg from '../../../../assets/webm/爱弥斯动态图表_transparent.webm'
import greenThemeImg from '../../../../assets/webm/卡提西娅动态图标_clean_edge.webm'
import purpleThemeImg from '../../../../assets/webm/椿动态图标视频_transparent.webm'
import darkDefaultImg from '../../../../assets/webm/星炬学院（带文字） 动态图标_transparent.webm'
import darkGreenImg from '../../../../assets/webm/先行公约动态图标_transparent.webm'
import darkPurpleImg from '../../../../assets/webm/黑海岸动态图标-彩色_transparent.webm'
import {
  ChatDotRound,
  Check,
  Close,
  Microphone,
  Picture,
  Promotion,
  VideoCamera,
  VideoPause
} from '@element-plus/icons-vue'

const isDarkMode = ref(false)

const props = defineProps({
  currentUser: { type: Object, default: null },
  currentUserOnline: { type: Boolean, default: false },
  currentUserOnlineLoading: { type: Boolean, default: false },
  userInfo: { type: Object, default: null },
  currentMessages: { type: Array, required: true },
  emojis: { type: Array, required: true },
  inputText: { type: String, required: true },
  isRecording: { type: Boolean, default: false },
  formattedDuration: { type: String, default: '00:00' },
  playingUrl: { type: String, default: '' },
  themeColor: { type: String, default: 'default' },
  setMsgListRef: { type: Function, required: true }
})

const emit = defineEmits([
  'update:inputText',
  'start-call',
  'more-click',
  'preview-image',
  'play-voice',
  'emoji-click',
  'select-image',
  'select-video',
  'toggle-voice',
  'cancel-recording',
  'send-voice',
  'send-text'
])

onMounted(() => {
  isDarkMode.value = document.documentElement.classList.contains('dark')
  
  const observer = new MutationObserver(() => {
    isDarkMode.value = document.documentElement.classList.contains('dark')
  })
  
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  })
})

const inputProxy = computed({
  get: () => props.inputText,
  set: (val) => emit('update:inputText', val)
})

const welcomeImg = computed(() => {
  if (isDarkMode.value) {
    switch (props.themeColor) {
      case 'green':
        return darkGreenImg
      case 'purple':
        return darkPurpleImg
      default:
        return darkDefaultImg
    }
  }
  
  switch (props.themeColor) {
    case 'green':
      return greenThemeImg
    case 'purple':
      return purpleThemeImg
    default:
      return defaultThemeImg
  }
})

const isVideoWelcome = computed(() => {
  return true
})
</script>
