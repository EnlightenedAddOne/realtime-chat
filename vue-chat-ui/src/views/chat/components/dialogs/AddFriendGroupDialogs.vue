<template>
  <el-dialog
    :model-value="showAddFriendDialog"
    title="添加好友 / 群组"
    width="480px"
    class="custom-dialog"
    :close-on-click-modal="false"
    append-to-body
    @update:model-value="emit('update:showAddFriendDialog', $event)"
  >
    <el-tabs :model-value="activeAddTab" class="custom-tabs" @update:model-value="emit('update:activeAddTab', $event)">
      <el-tab-pane label="找人" name="people">
        <div class="dialog-search">
          <el-input
            :model-value="addFriendKeyword"
            placeholder="搜索用户名或昵称"
            prefix-icon="Search"
            class="dialog-search-input"
            @update:model-value="emit('update:addFriendKeyword', $event)"
            @keyup.enter="emit('search-users')"
          >
            <template #append>
              <el-button @click="emit('search-users')">搜索</el-button>
            </template>
          </el-input>
        </div>

        <div class="search-results custom-scrollbar" v-loading="loadingSearch">
          <div v-if="searchResults.length === 0 && hasSearched" class="no-results">未找到用户</div>
          <div v-for="user in searchResults" :key="user.id" class="result-item dialog-card-item">
            <el-avatar :src="user.avatar_url" :size="40" shape="square" class="card-avatar">
              {{ user.nickname?.[0] || user.username?.[0] || 'U' }}
            </el-avatar>
            <div class="result-info">
              <div class="result-name">{{ user.nickname || user.username }}</div>
              <div class="result-id">ID: {{ user.username }}</div>
            </div>

            <el-button v-if="user.is_friend" type="info" round size="small" disabled>已添加</el-button>
            <el-button v-else-if="user.has_pending" type="warning" round size="small" disabled>等待中</el-button>
            <el-button v-else type="primary" round size="small" @click="emit('send-friend-request', user)">添加</el-button>
          </div>
        </div>

        <div v-if="pendingRequests.length > 0" class="pending-section">
          <h4 class="section-title">好友请求</h4>
          <div v-for="req in pendingRequests" :key="req.id" class="result-item request-item dialog-card-item">
            <el-avatar :src="req.sender?.avatar_url" :size="40" shape="square" class="card-avatar">
              {{ req.sender?.nickname?.[0] || req.sender?.username?.[0] || 'U' }}
            </el-avatar>
            <div class="result-info">
              <div class="result-name">{{ req.sender?.nickname || req.sender?.username }}</div>
              <div class="result-id">{{ req.remark || '无备注' }}</div>
            </div>
            <div class="request-actions">
              <el-button type="success" circle size="small" icon="Check" @click="emit('handle-request', req.id, 'accept')" />
              <el-button type="danger" circle size="small" icon="Close" @click="emit('handle-request', req.id, 'reject')" />
            </div>
          </div>
        </div>

        <div v-if="pendingGroupInvites.length > 0" class="pending-section">
          <h4 class="section-title">群邀请</h4>
          <div v-for="invite in pendingGroupInvites" :key="invite.id" class="result-item request-item dialog-card-item">
            <el-avatar :src="invite.inviter?.avatar_url" :size="40" shape="square" class="card-avatar">
              {{ invite.inviter?.nickname?.[0] || invite.inviter?.username?.[0] || 'U' }}
            </el-avatar>
            <div class="result-info">
              <div class="result-name">{{ invite.inviter?.nickname || invite.inviter?.username }} 邀请你加入</div>
              <div class="result-id">{{ invite.group?.name || '群聊' }}</div>
            </div>
            <div class="request-actions">
              <el-button type="success" circle size="small" icon="Check" @click="emit('handle-group-invite', invite.id, 'accept')" />
              <el-button type="danger" circle size="small" icon="Close" @click="emit('handle-group-invite', invite.id, 'reject')" />
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="找群" name="groups">
        <div class="dialog-search">
          <el-input
            :model-value="groupSearchKeyword"
            placeholder="搜索群组名称"
            prefix-icon="Search"
            class="dialog-search-input"
            @update:model-value="emit('update:groupSearchKeyword', $event)"
            @keyup.enter="emit('search-groups')"
          >
            <template #append>
              <el-button @click="emit('search-groups')">搜索</el-button>
            </template>
          </el-input>
        </div>

        <div class="search-results custom-scrollbar" v-loading="loadingGroupSearch">
          <div v-if="groupSearchResults.length === 0 && hasSearchedGroup" class="no-results">未找到群组</div>
          <div v-for="group in groupSearchResults" :key="group.id" class="result-item dialog-card-item">
            <el-avatar :src="group.avatar" :size="40" shape="square" class="card-avatar">
              {{ group.name?.[0] || 'G' }}
            </el-avatar>
            <div class="result-info">
              <div class="result-name">{{ group.name }}</div>
              <div class="result-id">ID: {{ group.id }}</div>
            </div>

            <el-button v-if="group.is_member" type="info" round size="small" disabled>已加入</el-button>
            <el-button v-else-if="group.has_pending" type="warning" round size="small" disabled>申请中</el-button>
            <el-button v-else type="primary" round size="small" @click="emit('open-apply-group-dialog', group)">申请加入</el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>

  <el-dialog
    :model-value="showApplyGroupDialog"
    title="申请加入群组"
    width="400px"
    class="custom-dialog"
    append-to-body
    @update:model-value="emit('update:showApplyGroupDialog', $event)"
  >
    <div class="dialog-card-item">
      <el-input
        :model-value="applyGroupReason"
        placeholder="请输入申请理由/备注"
        maxlength="50"
        show-word-limit
        type="textarea"
        :rows="3"
        @update:model-value="emit('update:applyGroupReason', $event)"
      />
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button round @click="emit('update:showApplyGroupDialog', false)">取消</el-button>
        <el-button type="primary" round @click="emit('apply-group')">发送申请</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
defineProps({
  showAddFriendDialog: { type: Boolean, required: true },
  activeAddTab: { type: String, required: true },
  addFriendKeyword: { type: String, required: true },
  searchResults: { type: Array, required: true },
  loadingSearch: { type: Boolean, default: false },
  hasSearched: { type: Boolean, default: false },
  pendingRequests: { type: Array, required: true },
  pendingGroupInvites: { type: Array, required: true },
  groupSearchKeyword: { type: String, required: true },
  groupSearchResults: { type: Array, required: true },
  loadingGroupSearch: { type: Boolean, default: false },
  hasSearchedGroup: { type: Boolean, default: false },
  showApplyGroupDialog: { type: Boolean, required: true },
  applyGroupReason: { type: String, required: true }
})

const emit = defineEmits([
  'update:showAddFriendDialog',
  'update:activeAddTab',
  'update:addFriendKeyword',
  'search-users',
  'send-friend-request',
  'handle-request',
  'handle-group-invite',
  'update:groupSearchKeyword',
  'search-groups',
  'open-apply-group-dialog',
  'update:showApplyGroupDialog',
  'update:applyGroupReason',
  'apply-group'
])
</script>
