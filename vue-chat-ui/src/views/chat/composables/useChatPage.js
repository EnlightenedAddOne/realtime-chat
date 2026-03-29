import { ref, computed, nextTick, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '../../../store/user'
import { useChatStore } from '../../../store/chat'
import { useCallStore } from '../../../store/call'
import { sendMessage, closeWebSocket, connectWebSocket } from '../../../utils/ws'
import { EMOJIS } from '../constants/emojis'
import { useResizableSidebar } from './useResizableSidebar'
import { useVoiceRecorder } from './useVoiceRecorder'
import { useAudioPlayer } from './useAudioPlayer'
import { useFriendFeatures } from './useFriendFeatures'
import { useGroupDiscovery } from './useGroupDiscovery'
import api from '../../../api'
import { normalizeMessageMedia, resolveMediaURL } from '../../../utils/mediaUrl'

export function useChatPage() {
  const router = useRouter()
  const userStore = useUserStore()
  const chatStore = useChatStore()
  const callStore = useCallStore()

  const activeTab = ref('messages')
  const {
    sidebarWidth,
    startResize,
    onMouseMove,
    stopResize
  } = useResizableSidebar()

  const searchText = ref('')
  const inputText = ref('')
  const currentUser = ref(null)
  const showFriendInfo = ref(false)
  const currentUserOnlineLoading = ref(false)
  const msgListRef = ref(null)
  const showImageViewer = ref(false)
  const previewUrl = ref('')

  const emojis = EMOJIS

  function getDefaultGroupAvatar(groupID, groupName = '') {
    const seed = encodeURIComponent(`${groupName || 'group'}-${groupID || '0'}`)
    return `https://api.dicebear.com/7.x/shapes/svg?seed=${seed}&backgroundColor=b6e3f4,c0aede,d1d4f9`
  }

  function handleEmojiClick(emoji) {
    inputText.value += emoji
  }

  const {
    showAddFriendDialog,
    activeAddTab,
    addFriendKeyword,
    searchResults,
    loadingSearch,
    hasSearched,
    pendingRequests,
    pendingGroupInvites,
    deletingFriendIds,
    searchUsers,
    sendFriendRequest,
    loadPendingRequests,
    handleRequest,
    handleGroupInvite,
    removeFriend
  } = useFriendFeatures({
    api,
    ElMessage,
    userStore,
    loadFriends,
    refreshConversations: () => chatStore.getConversations()
  })

  async function handleDeleteFriend(item) {
    if (!item || item.isGroup) return
    try {
      await ElMessageBox.confirm(`确认删除好友 ${item.name || item.nickname || item.username || '该用户'} 吗？`, '删除好友', {
        type: 'warning'
      })
      const ok = await removeFriend(item.id)
      if (!ok) return
      if (currentUser.value && !currentUser.value.isGroup && currentUser.value.id === item.id) {
        currentUser.value = null
        chatStore.setCurrentChatUser(null)
        showFriendInfo.value = false
      }
    } catch (e) {
      if (e !== 'cancel' && e !== 'close') {
        ElMessage.error('删除好友失败')
      }
    }
  }

  const {
    showCreateGroupDialog,
    createGroupName,
    groupSearchKeyword,
    groupSearchResults,
    loadingGroupSearch,
    hasSearchedGroup,
    showApplyGroupDialog,
    applyGroupReason,
    handleCreateGroup,
    searchGroups,
    openApplyGroupDialog,
    handleApplyGroup
  } = useGroupDiscovery({
    api,
    ElMessage,
    refreshConversations: () => chatStore.getConversations()
  })

  const groupRequests = ref([])

  const showLogoutDialog = ref(false)

  function handleLogoutClick() {
    showLogoutDialog.value = true
  }

  function confirmLogout() {
    closeWebSocket()
    userStore.logout()
    showLogoutDialog.value = false
    router.push('/login')
  }

  onMounted(async () => {
    if (!userStore.token) {
      router.push('/login')
      return
    }

    if (!userStore.userInfo) {
      try {
        await userStore.getUserInfo()
      } catch (e) {
        console.error('Failed to load user info on mount', e)
      }
    }

    connectWebSocket()
    loadFriends()
    chatStore.getConversations()
    loadPendingRequests()
  })

  function formatTime(date) {
    if (!date) return ''
    const d = new Date(date)
    if (isNaN(d.getTime())) return ''

    const now = new Date()

    if (d.getDate() === now.getDate() && d.getMonth() === now.getMonth() && d.getFullYear() === now.getFullYear()) {
      return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    }

    const yesterday = new Date(now)
    yesterday.setDate(now.getDate() - 1)
    if (d.getDate() === yesterday.getDate() && d.getMonth() === yesterday.getMonth() && d.getFullYear() === yesterday.getFullYear()) {
      return '昨天'
    }

    return d.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
  }

  const sidebarList = computed(() => {
    let rawList = []

    if (activeTab.value === 'messages') {
      rawList = chatStore.conversations.map(c => {
        if (c.type === 2) {
          return {
            id: c.peer_id,
            isGroup: true,
            target: c.group,
            name: c.group?.name || '群组',
            avatar: resolveMediaURL(c.group?.avatar) || getDefaultGroupAvatar(c.peer_id, c.group?.name),
            lastMsg: c.last_msg_content,
            time: c.updated_at,
            unread: c.unread_count || 0
          }
        }
        return {
          id: c.peer_id,
          isGroup: false,
          target: c.peer,
          name: c.peer?.nickname || c.peer?.username || 'Unknown',
          avatar: resolveMediaURL(c.peer?.avatar_url),
          lastMsg: c.last_msg_content,
          time: c.updated_at,
          unread: c.unread_count || 0
        }
      })
    } else if (activeTab.value === 'contacts') {
      rawList = chatStore.friends.map(f => ({
        id: f.friend_id,
        isGroup: false,
        target: f.friend_user,
        name: f.friend_user?.nickname || f.friend_user?.username || 'Unknown',
        avatar: resolveMediaURL(f.friend_user?.avatar_url),
        lastMsg: f.last_msg || (f.friend_user?.signature || ''),
        time: f.last_msg_time,
        unread: 0
      }))
    } else if (activeTab.value === 'groups') {
      rawList = chatStore.conversations
        .filter(c => c.type === 2)
        .map(c => ({
          id: c.peer_id,
          isGroup: true,
          target: c.group,
          name: c.group?.name || '群组',
          avatar: resolveMediaURL(c.group?.avatar) || getDefaultGroupAvatar(c.peer_id, c.group?.name),
          lastMsg: c.last_msg_content,
          time: c.updated_at,
          unread: c.unread_count || 0
        }))
    }

    if (searchText.value) {
      const lower = searchText.value.toLowerCase()
      rawList = rawList.filter(item => item.name.toLowerCase().includes(lower))
    }

    if (activeTab.value === 'messages' || activeTab.value === 'groups') {
      rawList.sort((a, b) => {
        const tA = new Date(a.time).getTime() || 0
        const tB = new Date(b.time).getTime() || 0
        return tB - tA
      })
    }

    return rawList
  })

  const currentMessages = computed(() => {
    if (!currentUser.value) return []
    const key = currentUser.value.isGroup ? `group_${currentUser.value.id}` : `user_${currentUser.value.id}`
    return chatStore.messages[key] || []
  })

  watch(currentMessages, () => {
    nextTick(() => {
      if (msgListRef.value) {
        msgListRef.value.scrollTop = msgListRef.value.scrollHeight
      }
      if (currentUser.value) {
        chatStore.markAsRead(currentUser.value.id, currentUser.value.isGroup ? 2 : 1)
      }
    })
  }, { deep: true })

  async function loadFriends() {
    try {
      const res = await api.get('/api/friends')
      if (res.data) {
        chatStore.setFriends(res.data)
      }
    } catch (e) {
      console.error('Failed to load friends:', e)
    }
  }

  async function selectChat(item) {
    const displayAvatar = item.isGroup
      ? (item.avatar || getDefaultGroupAvatar(item.id, item.name))
      : item.avatar

    const userObj = {
      id: item.id,
      isGroup: item.isGroup,
      nickname: item.name,
      username: item.name,
      avatar_url: resolveMediaURL(displayAvatar),
      owner_id: item.target?.owner_id,
      announcement: item.target?.announcement
    }

    currentUser.value = userObj
    chatStore.setCurrentChatUser(userObj)

    if (!item.isGroup) {
      await loadCurrentUserOnlineStatus(item.id, { force: true })
    } else {
      currentUserOnlineLoading.value = false
    }

    try {
      const params = item.isGroup ? { group_id: item.id } : { friend_id: item.id }
      const res = await api.get('/api/messages', { params })

      if (res.data) {
        const key = item.isGroup ? `group_${item.id}` : `user_${item.id}`
        chatStore.messages[key] = res.data.map(normalizeMessageMedia)
      }
    } catch (e) {
      console.error('Failed to load messages:', e)
    }
  }

  const currentUserOnline = computed(() => {
    const id = currentUser.value?.id
    if (!id || currentUser.value?.isGroup) return false
    const isOnline = !!chatStore.onlineStatus?.[id]
    console.log('[DEBUG] currentUserOnline computed: user', id, 'isOnline=', isOnline, 'onlineStatus=', chatStore.onlineStatus)
    return isOnline
  })

  async function loadCurrentUserOnlineStatus(targetUserID, options = {}) {
    const { force = false } = options
    if (!targetUserID) return

    if (!force && chatStore.onlineStatus?.[targetUserID] !== undefined) {
      return
    }

    currentUserOnlineLoading.value = true
    try {
      const res = await api.get('/api/user/online', {
        params: { user_id: targetUserID }
      })
      console.log('[DEBUG] loadCurrentUserOnlineStatus response for user', targetUserID, ':', res)
      console.log('[DEBUG] is_online value:', res?.is_online)
      chatStore.setUserOnlineStatus(targetUserID, !!res?.is_online)
      console.log('[DEBUG] chatStore.onlineStatus after setUserOnlineStatus:', JSON.stringify(chatStore.onlineStatus))
      console.log('[DEBUG] check user', targetUserID, 'in store:', chatStore.onlineStatus[targetUserID])
    } catch (e) {
      console.error('Failed to load user online status:', e)
    } finally {
      currentUserOnlineLoading.value = false
    }
  }

  function sendText() {
    if (!inputText.value.trim() || !currentUser.value) return

    const msg = {
      type: 'message',
      msg_type: 1,
      content: inputText.value.trim()
    }

    if (currentUser.value.isGroup) {
      msg.group_id = currentUser.value.id
    } else {
      msg.receiver_id = currentUser.value.id
    }

    if (sendMessage(msg)) {
      chatStore.addMessage(currentUser.value.id, {
        ...msg,
        sender_id: userStore.userInfo?.id,
        created_at: new Date().toISOString()
      }, currentUser.value.isGroup)
      inputText.value = ''
    }
  }

  function handleSelectImage() {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    input.onchange = async (e) => {
      const file = e.target.files[0]
      if (!file) return

      const formData = new FormData()
      formData.append('file', file)

      try {
        const res = await api.post('/api/upload', formData)
        const msg = {
          type: 'message',
          msg_type: 2,
          content: resolveMediaURL(res.url)
        }

        if (currentUser.value.isGroup) {
          msg.group_id = currentUser.value.id
        } else {
          msg.receiver_id = currentUser.value.id
        }

        if (sendMessage(msg)) {
          chatStore.addMessage(currentUser.value.id, {
            ...msg,
            sender_id: userStore.userInfo?.id,
            created_at: new Date().toISOString()
          }, currentUser.value.isGroup)
        }
      } catch (e) {
        ElMessage.error('图片上传失败')
      }
    }
    input.click()
  }

  function handleSelectVideo() {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'video/*'
    input.onchange = async (e) => {
      const file = e.target.files[0]
      if (!file) return

      const formData = new FormData()
      formData.append('file', file)

      try {
        const res = await api.post('/api/upload', formData)
        const msg = {
          type: 'message',
          msg_type: 4,
          content: resolveMediaURL(res.url)
        }

        if (currentUser.value.isGroup) {
          msg.group_id = currentUser.value.id
        } else {
          msg.receiver_id = currentUser.value.id
        }

        if (sendMessage(msg)) {
          chatStore.addMessage(currentUser.value.id, {
            ...msg,
            sender_id: userStore.userInfo?.id,
            created_at: new Date().toISOString()
          }, currentUser.value.isGroup)
        }
      } catch (e) {
        ElMessage.error('视频上传失败')
      }
    }
    input.click()
  }

  function getFileNameFromURL(url, fallback = 'download') {
    try {
      const parsed = new URL(url)
      const parts = parsed.pathname.split('/')
      const name = decodeURIComponent(parts[parts.length - 1] || '')
      return name || fallback
    } catch {
      const parts = (url || '').split('/')
      return decodeURIComponent(parts[parts.length - 1] || fallback)
    }
  }

  async function downloadByURL(url, fallbackName = 'download') {
    const normalizedURL = resolveMediaURL(url)
    if (!normalizedURL) return

    try {
      const response = await fetch(normalizedURL)
      if (!response.ok) {
        throw new Error(`download failed: ${response.status}`)
      }

      const blob = await response.blob()
      const objectURL = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = objectURL
      link.download = getFileNameFromURL(normalizedURL, fallbackName)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(objectURL)
    } catch (e) {
      console.error(e)
      ElMessage.error('下载失败')
    }
  }

  async function handleDownloadMessage(msg) {
    if (!msg?.content) return
    const defaultName = msg.msg_type === 2 ? 'image' : msg.msg_type === 4 ? 'video' : 'file'
    await downloadByURL(msg.content, defaultName)
  }

  function handleSelectFile() {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar,.7z,.mp3,.wav,.ogg,.flac,.jpg,.jpeg,.png,.gif,.bmp,.webp,.mp4,.webm,.mov,.avi,.mkv'
    input.onchange = async (e) => {
      const file = e.target.files[0]
      if (!file) return

      const formData = new FormData()
      formData.append('file', file)

      try {
        const res = await api.post('/api/upload', formData)
        const msg = {
          type: 'message',
          msg_type: 5,
          content: resolveMediaURL(res.url),
          file_name: res.original_name || file.name,
          file_size: file.size
        }

        if (currentUser.value.isGroup) {
          msg.group_id = currentUser.value.id
        } else {
          msg.receiver_id = currentUser.value.id
        }

        if (sendMessage(msg)) {
          chatStore.addMessage(currentUser.value.id, {
            ...msg,
            sender_id: userStore.userInfo?.id,
            created_at: new Date().toISOString()
          }, currentUser.value.isGroup)
        }
      } catch (e) {
        ElMessage.error('文件上传失败')
      }
    }
    input.click()
  }

  const {
    isRecording,
    formattedDuration,
    cancelRecording,
    toggleVoice,
    stopAndGetBlob
  } = useVoiceRecorder((message) => ElMessage.error(message))

  async function sendVoice() {
    if (!currentUser.value) return

    try {
      const { blob, duration } = await stopAndGetBlob()
      const file = new File([blob], `voice_${Date.now()}.webm`, { type: 'audio/webm' })
      const formData = new FormData()
      formData.append('file', file)

      const res = await api.post('/api/upload', formData)

      const msg = {
        type: 'message',
        msg_type: 3,
        content: resolveMediaURL(res.url),
        duration
      }

      if (currentUser.value.isGroup) {
        msg.group_id = currentUser.value.id
      } else {
        msg.receiver_id = currentUser.value.id
      }

      if (sendMessage(msg)) {
        chatStore.addMessage(currentUser.value.id, {
          ...msg,
          sender_id: userStore.userInfo?.id,
          created_at: new Date().toISOString()
        }, currentUser.value.isGroup)
      }
    } catch (e) {
      console.error(e)
      ElMessage.error('语音发送失败')
    }
  }

  const { playingUrl, playVoice } = useAudioPlayer((message) => ElMessage.error(message))

  function previewImage(url) {
    previewUrl.value = resolveMediaURL(url)
    showImageViewer.value = true
  }

  async function loadGroupRequests() {
    if (!currentUser.value?.id) return
    try {
      const res = await api.get(`/api/groups/${currentUser.value.id}/requests`)
      groupRequests.value = res.data || []
    } catch (e) {
      console.error('Failed to load group requests', e)
    }
  }

  async function handleGroupRequest(reqId, action) {
    try {
      await api.post('/api/groups/requests/handle', {
        req_id: reqId,
        action
      })
      ElMessage.success('操作成功')
      await loadGroupRequests()
      if (action === 'accept') {
        await loadGroupMembers()
      }
    } catch (e) {
      ElMessage.error('操作失败')
    }
  }

  
  
  const showAppConfigDialog = ref(false)
  const themeMode = ref(localStorage.getItem('chat-theme-mode') || 'system')
  const themeColor = ref(localStorage.getItem('chat-theme-color') || 'default')

  const applyThemeColor = (val) => {
    localStorage.setItem('chat-theme-color', val)
    const root = document.documentElement

    // Theme color only controls palette identity; light/dark surface variables are controlled by CSS selectors.
    if (val === 'green' || val === 'purple') {
      root.setAttribute('data-theme-color', val)
      return
    }

    root.setAttribute('data-theme-color', 'default')
  }

  const applyTheme = (val) => {
    localStorage.setItem('chat-theme-mode', val)
    const preferDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
    if (val === 'dark' || (val === 'system' && preferDark)) {
      document.documentElement.classList.add('dark')
      // Update body background
      document.body.style.backgroundColor = '#0f172a'
    } else {
      document.documentElement.classList.remove('dark')
      // Reset body background
      document.body.style.backgroundColor = '#f5f5f5'
    }
  }

  onMounted(() => {
    // Apply theme mode (light/dark)
    applyTheme(themeMode.value)
    // Apply theme color (default/green/purple)
    applyThemeColor(themeColor.value)
    
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
      if (themeMode.value === 'system') {
        applyTheme('system')
      }
    })
  })

  const showSettingsDialog = ref(false)




  const settingsForm = ref({
    nickname: '',
    avatar_url: ''
  })
  const updatingProfileAvatar = ref(false)

  const showGroupInfo = ref(false)
  const groupMembers = ref([])
  const loadingMembers = ref(false)
  const showAddMemberDialog = ref(false)
  const selectedFriends = ref([])
  const addingMembers = ref(false)
  const editingAnnouncement = ref(false)
  const announcementInput = ref('')
  const groupNameInput = ref('')
  const updatingGroupAvatar = ref(false)

    const showCropDialog = ref(false)
    const cropImageUrl = ref('')
    const cropTarget = ref('') // 'profile' | 'group'
    const isUploadingCropped = ref(false)

  const myGroupRole = computed(() => {
    if (!currentUser.value?.isGroup || !userStore.userInfo?.id) return 0
    const me = groupMembers.value.find(m => m.id === userStore.userInfo.id)
    return me?.role || 0
  })

  const isGroupOwner = computed(() => {
    return currentUser.value?.isGroup && currentUser.value?.owner_id === userStore.userInfo?.id
  })
  const isGroupAdmin = computed(() => myGroupRole.value === 2)
  const isGroupManager = computed(() => myGroupRole.value >= 2)

  const friendsNotInGroup = computed(() => {
    if (!currentUser.value?.isGroup) return []
    const memberIds = new Set(groupMembers.value.map(m => m.id))
    return chatStore.friends.filter(f => !memberIds.has(f.friend_id))
  })

  async function openGroupInfo() {
    if (!currentUser.value?.isGroup) return
    showGroupInfo.value = true
    await loadGroupMembers()
    announcementInput.value = currentUser.value?.announcement || ''
    groupNameInput.value = currentUser.value?.nickname || ''
    if (isGroupManager.value) {
      loadGroupRequests()
    }
  }

  function handleMoreClick() {
    if (currentUser.value?.isGroup) {
      openGroupInfo()
    } else {
      showFriendInfo.value = true
    }
  }

  const localVideoRef = ref(null)
  const remoteVideoRef = ref(null)

  watch(localVideoRef, (el) => {
    if (el && callStore.localStream) {
      el.srcObject = callStore.localStream
    }
  })

  watch(remoteVideoRef, (el) => {
    if (el && callStore.remoteStream) {
      el.srcObject = callStore.remoteStream
    }
  })

  watch(() => callStore.localStream, (stream) => {
    if (localVideoRef.value) {
      localVideoRef.value.srcObject = stream
    }
  })

  watch(() => callStore.remoteStream, (stream) => {
    if (remoteVideoRef.value) {
      remoteVideoRef.value.srcObject = stream
    }
  })

  function handleStartCall() {
    if (currentUser.value && !currentUser.value.isGroup) {
      callStore.startCall(currentUser.value)
    }
  }

  async function loadGroupMembers() {
    if (!currentUser.value?.id) return
    loadingMembers.value = true
    try {
      const res = await api.get(`/api/groups/${currentUser.value.id}/members`)
      groupMembers.value = res.data || []
    } catch (e) {
      ElMessage.error('获取群成员失败')
    } finally {
      loadingMembers.value = false
    }
  }

  function openAddMemberDialog() {
    selectedFriends.value = []
    showAddMemberDialog.value = true
  }

  async function confirmAddMembers() {
    if (selectedFriends.value.length === 0) return

    addingMembers.value = true
    let successCount = 0
    let failCount = 0

    try {
      const promises = selectedFriends.value.map(friendId =>
        api.post('/api/groups/join', {
          group_id: currentUser.value.id,
          user_id: friendId
        }).then(() => {
          successCount++
        }).catch(() => {
          failCount++
        })
      )

      await Promise.all(promises)

      if (successCount > 0) {
        ElMessage.success(`成功邀请 ${successCount} 位好友`)
        await loadGroupMembers()
        showAddMemberDialog.value = false
      }
      if (failCount > 0) {
        ElMessage.warning(`${failCount} 位好友邀请失败`)
      }
    } catch (e) {
      ElMessage.error('操作异常')
    } finally {
      addingMembers.value = false
    }
  }

  async function saveGroupAnnouncement() {
    if (!currentUser.value?.isGroup || !isGroupOwner.value) return
    try {
      await api.patch(`/api/groups/${currentUser.value.id}/announcement`, {
        announcement: announcementInput.value || ''
      })
      currentUser.value.announcement = announcementInput.value
      ElMessage.success('群公告已更新')
      editingAnnouncement.value = false
      await chatStore.getConversations()
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '更新群公告失败')
    }
  }

  async function saveGroupName() {
    if (!currentUser.value?.isGroup || !isGroupManager.value) return
    if (!groupNameInput.value.trim()) {
      ElMessage.warning('群名称不能为空')
      return
    }
    try {
      await api.patch(`/api/groups/${currentUser.value.id}/info`, {
        name: groupNameInput.value.trim()
      })
      currentUser.value.nickname = groupNameInput.value.trim()
      ElMessage.success('群名称已更新')
      await chatStore.getConversations()
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '更新群名称失败')
    }
  }

  function updateGroupAvatar() {
    if (!currentUser.value?.isGroup || !isGroupManager.value || updatingGroupAvatar.value) return

    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    input.onchange = (e) => {
      const file = e.target.files[0]
      if (!file) return

      cropImageUrl.value = URL.createObjectURL(file)
      cropTarget.value = 'group'
      showCropDialog.value = true
    }
    input.click()
  }

  async function handleCropConfirm(blob) {
    if (!blob) return
    const formData = new FormData()
    formData.append('file', blob, 'avatar.jpg')
    isUploadingCropped.value = true

    try {
      if (cropTarget.value === 'profile') {
        updatingProfileAvatar.value = true
      } else {
        updatingGroupAvatar.value = true
      }

      const uploadRes = await api.post('/api/upload', formData)
      const avatarURL = resolveMediaURL(uploadRes.url)
      if (!avatarURL) throw new Error('missing upload url')

      if (cropTarget.value === 'profile') {
        settingsForm.value.avatar_url = avatarURL
        ElMessage.success('头像上传成功，请保存设置')
      } else if (cropTarget.value === 'group') {
        const patchRes = await api.patch(`/api/groups/${currentUser.value.id}/avatar`, {
          avatar: avatarURL
        })
        const savedAvatar = patchRes.data?.avatar || avatarURL
        
        currentUser.value.avatar_url = savedAvatar
        ElMessage.success('群头像已更新')
        await chatStore.getConversations()
      }
      showCropDialog.value = false
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '头像更新失败')
    } finally {
      isUploadingCropped.value = false
      if (cropTarget.value === 'profile') {
        updatingProfileAvatar.value = false
      } else {
        updatingGroupAvatar.value = false
      }
    }
  }

  function canRemoveMember(member) {
    if (!member || !currentUser.value?.isGroup || member.id === userStore.userInfo?.id) return false
    if (isGroupOwner.value) return member.role !== 3
    if (isGroupAdmin.value) return member.role === 1
    return false
  }

  function canSetAdmin(member) {
    return isGroupOwner.value && member?.role === 1 && member.id !== userStore.userInfo?.id
  }

  function canUnsetAdmin(member) {
    return isGroupOwner.value && member?.role === 2 && member.id !== userStore.userInfo?.id
  }

  async function removeGroupMember(member) {
    if (!canRemoveMember(member)) return
    try {
      await ElMessageBox.confirm(`确认将 ${member.nickname || member.username} 移出群聊吗？`, '删除成员', {
        type: 'warning'
      })
      await api.post('/api/groups/members/remove', {
        group_id: currentUser.value.id,
        user_id: member.id
      })
      ElMessage.success('成员已移除')
      await loadGroupMembers()
    } catch (e) {
      if (e !== 'cancel' && e !== 'close') {
        ElMessage.error(e.response?.data?.message || '删除成员失败')
      }
    }
  }

  async function addGroupAdmin(member) {
    if (!canSetAdmin(member)) return
    try {
      await api.post('/api/groups/admins/add', {
        group_id: currentUser.value.id,
        user_id: member.id
      })
      ElMessage.success('已设为管理员')
      await loadGroupMembers()
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '设置管理员失败')
    }
  }

  async function removeGroupAdmin(member) {
    if (!canUnsetAdmin(member)) return
    try {
      await api.post('/api/groups/admins/remove', {
        group_id: currentUser.value.id,
        user_id: member.id
      })
      ElMessage.success('已取消管理员')
      await loadGroupMembers()
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '取消管理员失败')
    }
  }

  async function transferGroupOwnership(member) {
    if (!isGroupOwner.value || !member || member.id === userStore.userInfo?.id) return
    try {
      await ElMessageBox.confirm(`确认将群主转让给 ${member.nickname || member.username} 吗？`, '转让群主', {
        type: 'warning'
      })
      await api.post('/api/groups/transfer', {
        group_id: currentUser.value.id,
        user_id: member.id
      })
      ElMessage.success('群主转让成功')
      currentUser.value.owner_id = member.id
      await loadGroupMembers()
      await chatStore.getConversations()
    } catch (e) {
      if (e !== 'cancel' && e !== 'close') {
        ElMessage.error(e.response?.data?.message || '群主转让失败')
      }
    }
  }

  async function dismissCurrentGroup() {
    if (!isGroupOwner.value || !currentUser.value?.id) return
    try {
      await ElMessageBox.confirm('确认解散该群聊吗？解散后不可恢复。', '解散群聊', {
        type: 'warning',
        confirmButtonText: '确认解散'
      })
      await api.post(`/api/groups/${currentUser.value.id}/dismiss`)
      ElMessage.success('群已解散')
      showGroupInfo.value = false
      currentUser.value = null
      chatStore.setCurrentChatUser(null)
      await chatStore.getConversations()
    } catch (e) {
      if (e !== 'cancel' && e !== 'close') {
        ElMessage.error(e.response?.data?.message || '解散群失败')
      }
    }
  }

  async function leaveCurrentGroup() {
  if (isGroupOwner.value || !currentUser.value?.id) return
  try {
    await ElMessageBox.confirm('确认退出该群聊吗？', '退出群聊', {
    type: 'warning',
    confirmButtonText: '确认退出'
    })
    await api.post(`/api/groups/${currentUser.value.id}/leave`)
    ElMessage.success('已退出群聊')
    showGroupInfo.value = false
    currentUser.value = null
    chatStore.setCurrentChatUser(null)
    await chatStore.getConversations()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
    ElMessage.error(e.response?.data?.message || '退群失败')
    }
  }
  }

  watch(currentUser, (newVal, oldVal) => {
    if (newVal?.id !== oldVal?.id) {
      showGroupInfo.value = false
      showFriendInfo.value = false
    }
  })

  watch(showSettingsDialog, (val) => {
    if (val) {
      settingsForm.value = {
        nickname: userStore.userInfo?.nickname || '',
        avatar_url: userStore.userInfo?.avatar_url || ''
      }
    }
  })

  function selectProfileAvatarFile() {
    if (updatingProfileAvatar.value) return

    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
      input.onchange = (e) => {
        const file = e.target.files?.[0]
        if (!file) return

        cropImageUrl.value = URL.createObjectURL(file)
        cropTarget.value = 'profile'
        showCropDialog.value = true
      }
      input.click()
    }

  async function handleUpdateProfile() {
    try {
      const res = await api.post('/api/user/profile', settingsForm.value)
      const user = res?.user || res?.data?.user

      if (!user) {
        ElMessage.error('更新失败：响应数据异常')
        return
      }

      userStore.setUserInfo(user)
      ElMessage.success('更新成功')
      showSettingsDialog.value = false
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '更新失败')
    }
  }

  return {
    userStore,
    chatStore,
    callStore,
    activeTab,
    sidebarWidth,
    startResize,
    onMouseMove,
    stopResize,
    searchText,
    inputText,
    currentUser,
    currentUserOnline,
    currentUserOnlineLoading,
    msgListRef,
    showImageViewer,
    previewUrl,
    emojis,
    handleEmojiClick,
    showAddFriendDialog,
    activeAddTab,
    addFriendKeyword,
    searchResults,
    loadingSearch,
    hasSearched,
    pendingRequests,
    pendingGroupInvites,
    deletingFriendIds,
    searchUsers,
    sendFriendRequest,
    handleRequest,
    handleGroupInvite,
    handleDeleteFriend,
    showCreateGroupDialog,
    createGroupName,
    groupSearchKeyword,
    groupSearchResults,
    loadingGroupSearch,
    hasSearchedGroup,
    showApplyGroupDialog,
    applyGroupReason,
    handleCreateGroup,
    searchGroups,
    openApplyGroupDialog,
    handleApplyGroup,
    groupRequests,
    showLogoutDialog,
    handleLogoutClick,
    confirmLogout,
    formatTime,
    sidebarList,
    currentMessages,
    selectChat,
    sendText,
    handleSelectImage,
    handleSelectVideo,
    handleSelectFile,
    isRecording,
    formattedDuration,
    cancelRecording,
    toggleVoice,
    sendVoice,
    playingUrl,
    playVoice,
    previewImage,
    handleDownloadMessage,
    getFileNameFromURL,
    loadGroupRequests,
    handleGroupRequest,
    showSettingsDialog,
    showAppConfigDialog,
    themeMode,
    themeColor,
    applyTheme,
    applyThemeColor,
    settingsForm,
    updatingProfileAvatar,
    showGroupInfo,
    showFriendInfo,
    groupMembers,
    loadingMembers,
    showAddMemberDialog,
    selectedFriends,
    addingMembers,
    editingAnnouncement,
    announcementInput,
    groupNameInput,
    updatingGroupAvatar,
    isGroupOwner,
    isGroupAdmin,
    isGroupManager,
    myGroupRole,
    friendsNotInGroup,
    openGroupInfo,
    saveGroupAnnouncement,
    saveGroupName,
    updateGroupAvatar,
    canRemoveMember,
    canSetAdmin,
    canUnsetAdmin,
    removeGroupMember,
    addGroupAdmin,
    removeGroupAdmin,
    transferGroupOwnership,
    dismissCurrentGroup,
    leaveCurrentGroup,
    handleMoreClick,
    localVideoRef,
    remoteVideoRef,
    handleStartCall,
    openAddMemberDialog,
    confirmAddMembers,
    selectProfileAvatarFile,
      handleUpdateProfile,
      showCropDialog,
      cropImageUrl,
      isUploadingCropped,
      handleCropConfirm
    }
}
