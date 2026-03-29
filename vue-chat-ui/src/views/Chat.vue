<template>
  <div class="app-background" @mouseup="stopResize" @mousemove="onMouseMove">
    <!-- Main Full Screen Window -->
    <div class="chat-window">
      
      <MiniSidebar
        :user-info="userStore.userInfo"
        :active-tab="activeTab"
        @update:active-tab="activeTab = $event"
        @open-profile="showSettingsDialog = true"
        @open-app-config="showAppConfigDialog = true"
        @logout="handleLogoutClick"
      />

      <ConversationListSidebar
        :sidebar-width="sidebarWidth"
        :search-text="searchText"
        :sidebar-list="sidebarList"
        :current-user="currentUser"
        :active-tab="activeTab"
        :deleting-friend-ids="deletingFriendIds"
        :format-time="formatTime"
        @update:search-text="searchText = $event"
        @open-add-friend="showAddFriendDialog = true"
        @open-create-group="showCreateGroupDialog = true"
        @select-chat="selectChat"
        @delete-friend="handleDeleteFriend"
        @start-resize="startResize"
      />

      <ChatMainPanel
        :current-user="currentUser"
        :current-user-online="currentUserOnline"
        :current-user-online-loading="currentUserOnlineLoading"
        :user-info="userStore.userInfo"
        :current-messages="currentMessages"
        :emojis="emojis"
        :input-text="inputText"
        :is-recording="isRecording"
        :formatted-duration="formattedDuration"
        :playing-url="playingUrl"
        :theme-color="themeColor"
        :set-msg-list-ref="setMsgListRef"
        @update:input-text="inputText = $event"
        @start-call="handleStartCall"
        @more-click="handleMoreClick"
        @preview-image="previewImage"
        @play-voice="playVoice"
        @emoji-click="handleEmojiClick"
        @select-image="handleSelectImage"
        @select-video="handleSelectVideo"
        @toggle-voice="toggleVoice"
        @cancel-recording="cancelRecording"
        @send-voice="sendVoice"
        @send-text="sendText"
      />
    </div>

    <!-- Modals & Overlays -->
    <!-- Incoming Call Overlay -->
    <div v-if="callStore.incomingCall" class="incoming-call-overlay">
      <div class="incoming-call-card">
        <el-avatar :size="80" :src="callStore.incomingCall.sender_avatar" class="call-avatar">
          {{ callStore.incomingCall.sender_name?.[0] || 'U' }}
        </el-avatar>
        <div class="call-info">
          <h3>{{ callStore.incomingCall.sender_name }}</h3>
          <p>正在呼叫你...</p>
        </div>
        <div class="call-actions">
          <el-button type="danger" circle size="large" @click="callStore.rejectCall" class="action-btn hangup">
            <el-icon><Phone /></el-icon>
          </el-button>
          <el-button type="success" circle size="large" @click="callStore.acceptCall" class="action-btn answer">
            <el-icon><Phone /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <!-- Active Call Overlay -->
    <div v-if="callStore.isCalling" class="active-call-overlay">
      <!-- Remote Video (Main) -->
      <video ref="remoteVideoRef" autoplay playsinline class="remote-video"></video>
      
      <!-- Local Video (PIP) -->
      <div class="local-video-container">
        <video ref="localVideoRef" autoplay playsinline muted class="local-video"></video>
      </div>

      <!-- Controls -->
      <div class="call-controls">
         <div class="control-btn hangup" @click="callStore.hangup">
           <el-icon><Phone /></el-icon>
         </div>
      </div>
    </div>

    <el-image-viewer
      v-if="showImageViewer"
      :url-list="[previewUrl]"
      @close="showImageViewer = false"
    />

    <!-- Create Group Dialog -->
    <el-dialog
      v-model="showCreateGroupDialog"
      title="创建群组"
      width="400px"
      class="custom-dialog"
      append-to-body
    >
      <div class="dialog-card-item">
        <el-input 
          v-model="createGroupName" 
          placeholder="请输入群组名称" 
          maxlength="20"
          show-word-limit
        />
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateGroupDialog = false" round>取消</el-button>
          <el-button type="primary" @click="handleCreateGroup" round>创建</el-button>
        </span>
      </template>
    </el-dialog>

    <AddFriendGroupDialogs
      :show-add-friend-dialog="showAddFriendDialog"
      :active-add-tab="activeAddTab"
      :add-friend-keyword="addFriendKeyword"
      :search-results="searchResults"
      :loading-search="loadingSearch"
      :has-searched="hasSearched"
      :pending-requests="pendingRequests"
      :pending-group-invites="pendingGroupInvites"
      :group-search-keyword="groupSearchKeyword"
      :group-search-results="groupSearchResults"
      :loading-group-search="loadingGroupSearch"
      :has-searched-group="hasSearchedGroup"
      :show-apply-group-dialog="showApplyGroupDialog"
      :apply-group-reason="applyGroupReason"
      @update:show-add-friend-dialog="showAddFriendDialog = $event"
      @update:active-add-tab="activeAddTab = $event"
      @update:add-friend-keyword="addFriendKeyword = $event"
      @search-users="searchUsers"
      @send-friend-request="sendFriendRequest"
      @handle-request="handleRequest"
      @handle-group-invite="handleGroupInvite"
      @update:group-search-keyword="groupSearchKeyword = $event"
      @search-groups="searchGroups"
      @open-apply-group-dialog="openApplyGroupDialog"
      @update:show-apply-group-dialog="showApplyGroupDialog = $event"
      @update:apply-group-reason="applyGroupReason = $event"
      @apply-group="handleApplyGroup"
    />

    <ProfileSettingsDialog
      v-model="showSettingsDialog"
      :settings-form="settingsForm"
      :user-info="userStore.userInfo"
      :updating-profile-avatar="updatingProfileAvatar"
      @select-avatar="selectProfileAvatarFile"
      @save="handleUpdateProfile"
    />

    <AppConfigDialog
      v-model="showAppConfigDialog"
      :theme-mode="themeMode"
      :theme-color="themeColor"
      @update:theme-mode="themeMode = $event"
      @update:theme-color="themeColor = $event"
      @theme-change="applyTheme"
      @theme-color-change="applyThemeColor"
    />

    <GroupInfoDrawer
      v-model="showGroupInfo"
      :current-user="currentUser"
      :loading-members="loadingMembers"
      :is-group-manager="isGroupManager"
      :updating-group-avatar="updatingGroupAvatar"
      :group-name-input="groupNameInput"
      :editing-announcement="editingAnnouncement"
      :announcement-input="announcementInput"
      :group-requests="groupRequests"
      :group-members="groupMembers"
      :is-group-owner="isGroupOwner"
      :user-info="userStore.userInfo"
      :can-set-admin="canSetAdmin"
      :can-unset-admin="canUnsetAdmin"
      :can-remove-member="canRemoveMember"
      :show-add-member-dialog="showAddMemberDialog"
      :friends-not-in-group="friendsNotInGroup"
      :selected-friends="selectedFriends"
      :adding-members="addingMembers"
      @update-group-avatar="updateGroupAvatar"
      @update:group-name-input="groupNameInput = $event"
      @save-group-name="saveGroupName"
      @update:editing-announcement="editingAnnouncement = $event"
      @update:announcement-input="announcementInput = $event"
      @save-group-announcement="saveGroupAnnouncement"
      @handle-group-request="handleGroupRequest"
      @open-add-member-dialog="openAddMemberDialog"
      @add-group-admin="addGroupAdmin"
      @remove-group-admin="removeGroupAdmin"
      @transfer-group-ownership="transferGroupOwnership"
      @remove-group-member="removeGroupMember"
      @dismiss-current-group="dismissCurrentGroup"
      @update:show-add-member-dialog="showAddMemberDialog = $event"
      @update:selected-friends="selectedFriends = $event"
      @confirm-add-members="confirmAddMembers"
    />

    <!-- Logout Dialog -->
    <el-dialog
      v-model="showLogoutDialog"
      title="确认退出"
      width="320px"
      class="custom-dialog"
      append-to-body
    >
      <div class="dialog-card-item logout-content">
        <div class="logout-icon-bg">
          <el-icon class="logout-icon"><SwitchButton /></el-icon>
        </div>
        <span class="logout-text">确定要退出登录吗？</span>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showLogoutDialog = false" round>取消</el-button>
          <el-button type="danger" @click="confirmLogout" round>退出</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- Avatar Cropper -->
    <ImageCropper
      v-model="showCropDialog"
      :image-url="cropImageUrl"
      :loading="isUploadingCropped"
      @crop="handleCropConfirm"
    />
  </div>
</template>

<script setup>
import { 
  SwitchButton, 
  Phone,
  ChatLineRound
} from '@element-plus/icons-vue'
import { useChatPage } from './chat/composables/useChatPage'
import ImageCropper from '../components/ImageCropper.vue'
import MiniSidebar from './chat/components/layout/MiniSidebar.vue'
import ConversationListSidebar from './chat/components/layout/ConversationListSidebar.vue'
import ChatMainPanel from './chat/components/layout/ChatMainPanel.vue'
import ProfileSettingsDialog from './chat/components/dialogs/ProfileSettingsDialog.vue'
import AppConfigDialog from './chat/components/dialogs/AppConfigDialog.vue'
import AddFriendGroupDialogs from './chat/components/dialogs/AddFriendGroupDialogs.vue'
import GroupInfoDrawer from './chat/components/drawers/GroupInfoDrawer.vue'

const {
  userStore,
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
  isRecording,
  formattedDuration,
  cancelRecording,
  toggleVoice,
  sendVoice,
  playingUrl,
  playVoice,
  previewImage,
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
  friendsNotInGroup,
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
} = useChatPage()

function setMsgListRef(el) {
  msgListRef.value = el
}
</script>

<style src="./chat/styles/chat-scoped.css"></style>
<style src="./chat/styles/chat-global.css"></style>
<style src="./chat/styles/decorations.css"></style>
