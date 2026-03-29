<template>
  <el-drawer
    :model-value="modelValue"
    :title="currentUser?.nickname || '群组信息'"
    direction="rtl"
    size="360px"
    class="group-info-drawer"
    destroy-on-close
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="group-info-wrapper" v-loading="loadingMembers">
      <div class="group-hero">
        <el-avatar :size="80" :src="currentUser?.avatar_url" shape="circle" class="group-avatar shadow-avatar">
          {{ currentUser?.nickname?.[0] || '群' }}
        </el-avatar>
        <el-button
          v-if="isGroupManager"
          class="group-avatar-edit-btn"
          circle
          size="small"
          :loading="updatingGroupAvatar"
          title="修改群头像"
          @click="emit('update-group-avatar')"
        >
          <el-icon><Edit /></el-icon>
        </el-button>
        <h3 class="group-name-display">{{ currentUser?.nickname }}</h3>
        <p class="group-id-display">群号: {{ currentUser?.id }}</p>
      </div>

      <div class="group-body-scroll custom-scrollbar">
        <div class="group-card" v-if="isGroupManager">
          <div class="card-header">
            <el-icon class="card-icon text-primary"><Setting /></el-icon>
            <span class="card-title">群名称</span>
          </div>
          <div class="card-body">
            <el-input
              :model-value="groupNameInput"
              maxlength="50"
              placeholder="请输入新群名称"
              class="anime-input"
              @update:model-value="emit('update:groupNameInput', $event)"
            >
              <template #append>
                <el-button type="primary" class="save-append-btn" @click="emit('save-group-name')">修改</el-button>
              </template>
            </el-input>
          </div>
        </div>

        <div class="group-card">
          <div class="card-header flex-between">
            <div class="left-h">
              <el-icon class="card-icon text-warning"><ChatDotRound /></el-icon>
              <span class="card-title">群公告</span>
            </div>
            <el-button v-if="isGroupManager && !editingAnnouncement" size="small" type="primary" link @click="emit('update:editingAnnouncement', true)">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
          </div>
          <div class="card-body">
            <div v-if="isGroupManager && editingAnnouncement" class="announcement-edit-box">
              <el-input
                :model-value="announcementInput"
                type="textarea"
                :rows="3"
                maxlength="500"
                show-word-limit
                placeholder="说点什么吧..."
                class="anime-textarea"
                @update:model-value="emit('update:announcementInput', $event)"
              />
              <div class="action-row-right mt-2">
                <el-button size="small" round @click="emit('update:editingAnnouncement', false)">取消</el-button>
                <el-button size="small" type="primary" round @click="emit('save-group-announcement')">发布</el-button>
              </div>
            </div>
            <div v-else class="announcement-view">
              <p v-if="currentUser?.announcement">{{ currentUser.announcement }}</p>
              <p v-else class="text-placeholder">暂无群公告~</p>
            </div>
          </div>
        </div>

        <div class="group-card" v-if="isGroupManager && groupRequests.length > 0">
          <div class="card-header flex-between">
            <div class="left-h">
              <el-icon class="card-icon text-danger"><Bell /></el-icon>
              <span class="card-title">入群申请</span>
            </div>
            <el-badge :value="groupRequests.length" class="req-badge" type="danger" />
          </div>
          <div class="card-body p-0">
            <div class="requests-list">
              <div v-for="req in groupRequests" :key="req.id" class="req-item">
                <el-avatar :size="40" :src="req.user?.avatar_url" class="req-avatar">
                  {{ req.user?.nickname?.[0] || 'U' }}
                </el-avatar>
                <div class="req-info">
                  <div class="req-name">{{ req.user?.nickname || req.user?.username }}</div>
                  <div class="req-msg" :title="req.remark">{{ req.remark || '无备注' }}</div>
                </div>
                <div class="req-actions">
                  <el-button type="success" circle size="small" :icon="Check" @click="emit('handle-group-request', req.id, 'accept')"></el-button>
                  <el-button type="danger" circle plain size="small" :icon="Close" @click="emit('handle-group-request', req.id, 'reject')"></el-button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="group-card members-card">
          <div class="card-header flex-between">
            <div class="left-h">
              <el-icon class="card-icon text-success"><User /></el-icon>
              <span class="card-title">群成员 ({{ groupMembers.length }})</span>
            </div>
          </div>
          <div class="card-body p-0">
            <div class="members-grid custom-scrollbar">
              <div class="member-cell add-cell" @click="emit('open-add-member-dialog')">
                <div class="member-avatar-box dashed-box">
                  <el-icon><Plus /></el-icon>
                </div>
                <span class="member-name-txt">邀请</span>
              </div>

              <div v-for="member in groupMembers" :key="member.id" class="member-cell">
                <div class="member-avatar-box">
                  <el-avatar :size="44" :src="member.avatar_url" class="member-avatar">
                    {{ member.nickname?.[0] || 'U' }}
                  </el-avatar>
                  <div v-if="member.role === 3" class="role-tag owner">群主</div>
                  <div v-else-if="member.role === 2" class="role-tag admin">管理</div>

                  <div class="member-hover-menu" v-if="(isGroupOwner && member.id !== userInfo?.id) || (isGroupManager && member.role === 1 && member.id !== userInfo?.id)">
                    <el-dropdown trigger="click" placement="bottom">
                      <div class="hover-trigger">
                        <el-icon><More /></el-icon>
                      </div>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item v-if="canSetAdmin(member)" @click="emit('add-group-admin', member)">
                            <el-icon><Promotion /></el-icon>设为管理
                          </el-dropdown-item>
                          <el-dropdown-item v-if="canUnsetAdmin(member)" @click="emit('remove-group-admin', member)">
                            <el-icon><Remove /></el-icon>取消管理
                          </el-dropdown-item>
                          <el-dropdown-item v-if="isGroupOwner && member.id !== userInfo?.id && member.role !== 3" @click="emit('transfer-group-ownership', member)">
                            <el-icon><Switch /></el-icon>转让群主
                          </el-dropdown-item>
                          <el-dropdown-item v-if="canRemoveMember(member)" class="danger-item" divided @click="emit('remove-group-member', member)">
                            <el-icon><Delete /></el-icon>踢出该群
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </div>
                </div>
                <span class="member-name-txt truncate" :title="member.nickname || member.username">{{ member.nickname || member.username }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="group-card danger-card" v-if="isGroupOwner">
          <el-button type="danger" plain class="w-full" @click="emit('dismiss-current-group')">
            <el-icon><Warning /></el-icon>
            解散群聊
          </el-button>
        </div>
      </div>
    </div>
  </el-drawer>

  <el-dialog
    :model-value="showAddMemberDialog"
    title="邀请好友入群"
    width="400px"
    class="custom-dialog"
    append-to-body
    @update:model-value="emit('update:showAddMemberDialog', $event)"
  >
    <div class="select-friend-list custom-scrollbar">
      <el-checkbox-group :model-value="selectedFriends" @update:model-value="emit('update:selectedFriends', $event)">
        <div v-if="friendsNotInGroup.length === 0" class="no-friends-tip">暂无更多好友可邀请</div>
        <div v-for="friend in friendsNotInGroup" :key="friend.friend_id" class="friend-select-item">
          <el-checkbox :label="friend.friend_id" class="friend-checkbox">
            <div class="friend-info-row">
              <el-avatar :size="32" :src="friend.friend_user?.avatar_url">
                {{ friend.friend_user?.nickname?.[0] || 'U' }}
              </el-avatar>
              <span class="friend-name">{{ friend.friend_user?.nickname || friend.friend_user?.username }}</span>
            </div>
          </el-checkbox>
        </div>
      </el-checkbox-group>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button round @click="emit('update:showAddMemberDialog', false)">取消</el-button>
        <el-button type="primary" round :loading="addingMembers" :disabled="selectedFriends.length === 0" @click="emit('confirm-add-members')">
          确定邀请 ({{ selectedFriends.length }})
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import {
  Bell,
  ChatDotRound,
  Check,
  Close,
  Delete,
  Edit,
  More,
  Plus,
  Promotion,
  Remove,
  Setting,
  Switch,
  User,
  Warning
} from '@element-plus/icons-vue'

defineProps({
  modelValue: { type: Boolean, required: true },
  currentUser: { type: Object, default: null },
  loadingMembers: { type: Boolean, default: false },
  isGroupManager: { type: Boolean, default: false },
  updatingGroupAvatar: { type: Boolean, default: false },
  groupNameInput: { type: String, required: true },
  editingAnnouncement: { type: Boolean, default: false },
  announcementInput: { type: String, required: true },
  groupRequests: { type: Array, required: true },
  groupMembers: { type: Array, required: true },
  isGroupOwner: { type: Boolean, default: false },
  userInfo: { type: Object, default: null },
  canSetAdmin: { type: Function, required: true },
  canUnsetAdmin: { type: Function, required: true },
  canRemoveMember: { type: Function, required: true },
  showAddMemberDialog: { type: Boolean, required: true },
  friendsNotInGroup: { type: Array, required: true },
  selectedFriends: { type: Array, required: true },
  addingMembers: { type: Boolean, default: false }
})

const emit = defineEmits([
  'update:modelValue',
  'update-group-avatar',
  'update:groupNameInput',
  'save-group-name',
  'update:editingAnnouncement',
  'update:announcementInput',
  'save-group-announcement',
  'handle-group-request',
  'open-add-member-dialog',
  'add-group-admin',
  'remove-group-admin',
  'transfer-group-ownership',
  'remove-group-member',
  'dismiss-current-group',
  'update:showAddMemberDialog',
  'update:selectedFriends',
  'confirm-add-members'
])
</script>
