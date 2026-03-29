import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUserStore } from './user'
import api from '../api'
import { normalizeMessageMedia } from '../utils/mediaUrl'

export const useChatStore = defineStore('chat', () => {
  // 当前选中的聊天用户 (Peer Info)
  const currentChatUser = ref(null)
  
  // 消息列表 (key: peer_id, value: messages[])
  const messages = ref({})
  
  // 会话列表
  const conversations = ref([])
  
  // 好友列表
  const friends = ref([])

  // 在线状态映射 (key: userId, value: boolean)
  const onlineStatus = ref({})

  function setCurrentChatUser(user) {
    currentChatUser.value = user
    // Reset unread locally and sync to server
    if (user) {
      markAsRead(user.id, user.isGroup ? 2 : 1)
    }
  }

  async function getConversations() {
    try {
      const res = await api.get('/api/conversations')
      if (res.data) {
        conversations.value = res.data
      }
    } catch (e) {
      console.error('Failed to fetch conversations', e)
    }
  }

  async function markAsRead(peerId, type = 1) {
    // Optimistic update
    const conv = conversations.value.find(c => c.peer_id === peerId && c.type === type)
    if (conv) {
      conv.unread_count = 0
    }
    
    try {
      await api.post('/api/conversations/read', { peer_id: peerId, type })
    } catch (e) {
      console.error('Failed to mark as read', e)
    }
  }

  function setConversations(list) {
    conversations.value = list
  }
  
  function setFriends(list) {
    friends.value = list
  }

  function setUserOnlineStatus(userID, isOnline) {
    if (!userID) return
    onlineStatus.value[userID] = !!isOnline
  }

  function setOnlineSnapshot(onlineUserIDs = []) {
    const next = {}
    for (const id of onlineUserIDs) {
      if (id) next[id] = true
    }
    onlineStatus.value = next
  }

  // peerId: 对方的用户ID 或 群组ID
  // msg: 消息对象
  // isGroup: 是否群聊
  function addMessage(peerId, msg, isGroup = false) {
    const normalizedMsg = normalizeMessageMedia(msg)
    const userStore = useUserStore()
    const isSelf = normalizedMsg.sender_id === userStore.userInfo?.id
    
    // Construct key
    const key = isGroup ? `group_${peerId}` : `user_${peerId}`

    // 1. Add to messages map
    if (!messages.value[key]) {
      messages.value[key] = []
    }
    messages.value[key].push(normalizedMsg)

    // 2. Update conversation preview
    let conv = conversations.value.find(c => {
        if (isGroup) return c.type === 2 && c.peer_id === peerId
        return c.type === 1 && c.peer_id === peerId
    })
    
    // 生成摘要
    let preview = ''
    switch (normalizedMsg.msg_type) {
      case 1: preview = normalizedMsg.content; break;
      case 2: preview = '[图片]'; break;
      case 3: preview = '[语音]'; break;
      case 4: preview = '[视频]'; break;
      case 5: preview = '[文件]'; break;
      default: preview = '[消息]';
    }
    
    const now = new Date().toISOString()
    const currentTargetId = currentChatUser.value?.id
    // Check if current chat is the target
    // We need to know if currentChatUser is a group or user.
    // Let's assume currentChatUser has a 'type' property (1=user, 2=group) or we check existence of 'username' vs 'name'
    const isCurrentChat = currentChatUser.value && 
                          (isGroup ? (currentChatUser.value.isGroup && currentChatUser.value.id === peerId) 
                                   : (!currentChatUser.value.isGroup && currentChatUser.value.id === peerId))

    if (conv) {
      conv.last_msg_content = preview
      conv.updated_at = now
      
      // Update unread count
      if (!isSelf && !isCurrentChat) {
        conv.unread_count = (conv.unread_count || 0) + 1
      }

      // Move to top
      const idx = conversations.value.indexOf(conv)
      if (idx > 0) {
        conversations.value.splice(idx, 1)
        conversations.value.unshift(conv)
      }
    } else {
        // 如果会话不存在 (e.g. 新收到消息)，尝试创建或拉取
        // For groups, we might need to fetch the group info if we don't have it?
        // For now fallback to getConversations
        getConversations()
    }
  }

  return {
    currentChatUser,
    messages,
    conversations,
    friends,
    onlineStatus,
    setCurrentChatUser,
    setConversations,
    setFriends,
    setUserOnlineStatus,
    setOnlineSnapshot,
    addMessage,
    getConversations,
    markAsRead
  }
})
