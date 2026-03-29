import { useUserStore } from '../store/user'
import { useChatStore } from '../store/chat'
import { useCallStore } from '../store/call'
import { ElMessage } from 'element-plus'
import { config } from '../config/index.js'
import { normalizeMessageMedia } from './mediaUrl'

let ws = null
let reconnectTimer = null

export function connectWebSocket() {
  const userStore = useUserStore()
  const chatStore = useChatStore()
  const callStore = useCallStore()

  if (!userStore.token) {
    console.warn('WebSocket: 未登录，跳过连接')
    return
  }

  // 构建 WebSocket URL - 使用配置中的基础 URL
  const wsUrl = `${config.WS_BASE_URL}/ws?token=${userStore.token}`
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('✅ WebSocket 已连接')
    ElMessage.success('实时通讯已连接')
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      handleMessage(data, chatStore, callStore)
    } catch (e) {
      console.error('解析消息失败:', e)
    }
  }

  ws.onclose = () => {
    console.log('❌ WebSocket 已断开')
    // 自动重连
    reconnectTimer = setTimeout(() => {
      console.log('正在重连...')
      connectWebSocket()
    }, config.WS_RECONNECT_INTERVAL)
  }

  ws.onerror = (error) => {
    console.error('WebSocket 错误:', error)
  }
}

function handleMessage(data, chatStore, callStore) {
  const { type, ...payload } = data

  switch (type) {
    case 'signal':
      // The payload structure is { receiver_id, content: "{ "type": "offer", ... }" }
      // But the store expects the inner content mostly
      if (payload.content) {
          try {
              const signalData = JSON.parse(payload.content)
              callStore.handleSignal({
                  sender_id: payload.sender_id, // Ensure sender_id is passed
                  ...signalData
              })
          } catch (e) {
              console.error('Failed to parse signal content', e)
          }
      }
      break
    case 'message':
      // Determine if this is a group message by checking for group_id
      const normalizedPayload = normalizeMessageMedia(payload)
      const isGroup = !!normalizedPayload.group_id
      const peerId = isGroup ? normalizedPayload.group_id : normalizedPayload.sender_id
      chatStore.addMessage(peerId, normalizedPayload, isGroup)
      break
    case 'system':
      ElMessage.info(payload.content)
      break
    case 'presence':
      if (payload.target_user_id) {
        chatStore.setUserOnlineStatus(payload.target_user_id, !!payload.is_online)
      }
      break
    case 'presence_snapshot':
      chatStore.setOnlineSnapshot(payload.online_user_ids || [])
      break
    default:
      console.log('未知消息类型:', type)
  }
}

export function sendMessage(msg) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(msg))
    return true
  }
  console.error('WebSocket 未连接')
  return false
}

export function closeWebSocket() {
  if (ws) {
    ws.close()
    ws = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

export function getWebSocket() {
  return ws
}
