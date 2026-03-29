import { ref } from 'vue'

export function useFriendFeatures({ api, ElMessage, userStore, loadFriends, refreshConversations }) {
  const showAddFriendDialog = ref(false)
  const activeAddTab = ref('people')
  const addFriendKeyword = ref('')
  const searchResults = ref([])
  const loadingSearch = ref(false)
  const hasSearched = ref(false)
  const pendingRequests = ref([])
  const pendingGroupInvites = ref([])
  const deletingFriendIds = ref({})

  async function searchUsers() {
    if (!addFriendKeyword.value.trim()) return
    loadingSearch.value = true
    hasSearched.value = true
    try {
      const res = await api.get('/api/friends/search', {
        params: { keyword: addFriendKeyword.value }
      })
      searchResults.value = res.data || []
    } catch (e) {
      ElMessage.error('搜索失败')
    } finally {
      loadingSearch.value = false
    }
  }

  async function sendFriendRequest(user) {
    try {
      await api.post('/api/friends/request', {
        receiver_id: user.id,
        remark: `你好，我是 ${userStore.userInfo?.nickname || userStore.userInfo?.username}`
      })
      ElMessage.success('请求已发送')
      user.has_pending = true
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '请求发送失败')
    }
  }

  async function loadPendingRequests() {
    try {
      const [friendRes, groupInviteRes] = await Promise.all([
        api.get('/api/friends/requests'),
        api.get('/api/groups/invitations')
      ])
      pendingRequests.value = friendRes.data || []
      pendingGroupInvites.value = groupInviteRes.data || []
    } catch (e) {
      console.error('Failed to load pending requests')
    }
  }

  async function handleRequest(reqId, action) {
    try {
      await api.post('/api/friends/handle', {
        req_id: reqId,
        action
      })
      ElMessage.success(action === 'accept' ? '已接受' : '已拒绝')
      await loadPendingRequests()
      if (action === 'accept') {
        await loadFriends()
      }
    } catch (e) {
      ElMessage.error('操作失败')
    }
  }

  async function handleGroupInvite(reqId, action) {
    try {
      await api.post('/api/groups/invitations/handle', {
        req_id: reqId,
        action
      })
      ElMessage.success(action === 'accept' ? '已加入群聊' : '已拒绝群邀请')
      await loadPendingRequests()
      if (action === 'accept' && typeof refreshConversations === 'function') {
        await refreshConversations()
      }
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '操作失败')
    }
  }

  async function removeFriend(friendId) {
    if (!friendId) return false
    deletingFriendIds.value = { ...deletingFriendIds.value, [friendId]: true }
    try {
      await api.post('/api/friends/delete', { friend_id: friendId })
      ElMessage.success('已删除好友')
      await loadFriends()
      if (typeof refreshConversations === 'function') {
        await refreshConversations()
      }
      return true
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '删除好友失败')
      return false
    } finally {
      const next = { ...deletingFriendIds.value }
      delete next[friendId]
      deletingFriendIds.value = next
    }
  }

  return {
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
  }
}
