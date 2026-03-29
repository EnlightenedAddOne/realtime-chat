import { ref } from 'vue'

export function useGroupDiscovery({ api, ElMessage, refreshConversations }) {
  const showCreateGroupDialog = ref(false)
  const createGroupName = ref('')

  const groupSearchKeyword = ref('')
  const groupSearchResults = ref([])
  const loadingGroupSearch = ref(false)
  const hasSearchedGroup = ref(false)

  const showApplyGroupDialog = ref(false)
  const applyGroupReason = ref('')
  const currentApplyGroupId = ref(null)

  async function handleCreateGroup() {
    if (!createGroupName.value.trim()) return
    try {
      await api.post('/api/groups', { name: createGroupName.value.trim() })
      ElMessage.success('群组创建成功')
      showCreateGroupDialog.value = false
      createGroupName.value = ''
      await refreshConversations()
    } catch (e) {
      ElMessage.error('创建失败')
    }
  }

  async function searchGroups() {
    if (!groupSearchKeyword.value.trim()) return
    loadingGroupSearch.value = true
    hasSearchedGroup.value = true
    try {
      const res = await api.get('/api/search/groups', {
        params: { keyword: groupSearchKeyword.value }
      })
      groupSearchResults.value = res.data || []
    } catch (e) {
      ElMessage.error('搜索群组失败')
    } finally {
      loadingGroupSearch.value = false
    }
  }

  function openApplyGroupDialog(group) {
    currentApplyGroupId.value = group.id
    applyGroupReason.value = ''
    showApplyGroupDialog.value = true
  }

  async function handleApplyGroup() {
    if (!currentApplyGroupId.value) return
    try {
      await api.post('/api/groups/apply', {
        group_id: currentApplyGroupId.value,
        remark: applyGroupReason.value
      })
      ElMessage.success('申请已发送')
      showApplyGroupDialog.value = false
      const group = groupSearchResults.value.find(g => g.id === currentApplyGroupId.value)
      if (group) group.has_pending = true
    } catch (e) {
      ElMessage.error(e.response?.data?.message || '申请失败')
    }
  }

  return {
    showCreateGroupDialog,
    createGroupName,
    groupSearchKeyword,
    groupSearchResults,
    loadingGroupSearch,
    hasSearchedGroup,
    showApplyGroupDialog,
    applyGroupReason,
    currentApplyGroupId,
    handleCreateGroup,
    searchGroups,
    openApplyGroupDialog,
    handleApplyGroup
  }
}
