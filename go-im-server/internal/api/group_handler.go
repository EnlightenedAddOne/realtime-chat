package api

import (
	"fmt"
	"go-im-server/internal/model"
	"go-im-server/internal/repository"
	"go-im-server/pkg/logger"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GroupHandler struct {
	repo     *repository.GroupRepository
	convRepo *repository.ConversationRepository
}

func defaultGroupAvatar(seed string) string {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = "group"
	}
	return fmt.Sprintf("https://api.dicebear.com/7.x/shapes/svg?seed=%s&backgroundColor=b6e3f4,c0aede,d1d4f9", url.QueryEscape(seed))
}

func (h *GroupHandler) getOperatorRole(groupID, userID uint) (int8, error) {
	role, exists, err := h.repo.GetMemberRole(groupID, userID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}
	return role, nil
}

func NewGroupHandler(db *gorm.DB) *GroupHandler {
	return &GroupHandler{
		repo:     repository.NewGroupRepository(db),
		convRepo: repository.NewConversationRepository(db),
	}
}

// CreateGroup creates a new group with the current user as owner
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		Name         string `json:"name" binding:"required"`
		Avatar       string `json:"avatar"`
		Announcement string `json:"announcement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	// Create group instance with current user as owner
	avatar := strings.TrimSpace(req.Avatar)
	if avatar == "" {
		avatar = defaultGroupAvatar(req.Name)
	}

	group := &model.Group{
		OwnerID:      userID,
		Name:         req.Name,
		Avatar:       avatar,
		Announcement: req.Announcement,
		CreatedAt:    time.Now(),
	}

	// CreateGroup adds owner as member automatically
	if err := h.repo.CreateGroup(group, []uint{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建群组失败"})
		return
	}

	// Create conversation for owner so it appears in the list immediately
	// Use 0 as lastMsgID (no message yet)
	if err := h.convRepo.UpsertGroup(userID, group.ID, 0, "群组已创建", 0); err != nil {
		// Log error but don't fail the request, as group is already created
		// In a real app, we might want to use a transaction across services or compensating action
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "群组创建成功",
		"data":    group,
	})
}

// GetGroups returns all groups the current user belongs to
func (h *GroupHandler) GetGroups(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groups, err := h.repo.GetUserGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取群组列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": groups})
}

// GetGroupMembers returns all members of a group with full user details
func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群组ID错误"})
		return
	}

	isMember, err := h.repo.IsMember(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员状态失败"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"message": "非群成员无法查看成员列表"})
		return
	}

	members, err := h.repo.GetMemberDetailsWithRole(uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取成员列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": members})
}

// JoinGroup adds a member to the group
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	operatorID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		GroupID uint   `json:"group_id" binding:"required"`
		UserID  uint   `json:"user_id" binding:"required"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	group, err := h.repo.GetGroupInfo(req.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取群组信息失败"})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "群组不存在"})
		return
	}

	// Check if already a member
	isMember, err := h.repo.IsMember(req.GroupID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员状态失败"})
		return
	}
	if isMember {
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户已是群成员"})
		return
	}

	// 自己加群：仅允许自己直接加入（保留旧行为兼容）
	if req.UserID == operatorID {
		if err := h.repo.AddMember(req.GroupID, req.UserID, 1); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "加入群组失败"})
			return
		}

		if err := h.convRepo.UpsertGroup(req.UserID, req.GroupID, 0, "您已加入群聊", 0); err != nil {
			// Log error but don't fail the request
		}

		c.JSON(http.StatusOK, gin.H{"message": "加入群组成功"})
		return
	}

	// 邀请他人：群主或管理员允许邀请，且必须等待对方确认
	operatorRole, err := h.getOperatorRole(req.GroupID, operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if operatorRole < model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主或管理员可以邀请好友入群"})
		return
	}

	pendingInvite, err := h.repo.GetPendingGroupInvite(req.GroupID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查群邀请失败"})
		return
	}
	if pendingInvite != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "该好友已有待处理邀请"})
		return
	}

	inviterID := operatorID
	groupReq := &model.GroupRequest{
		UserID:    req.UserID,
		GroupID:   req.GroupID,
		InviterID: &inviterID,
		Remark:    req.Remark,
		Status:    model.GroupRequestStatusPending,
		Type:      model.GroupRequestTypeInvite,
	}
	if err := h.repo.CreateGroupRequest(groupReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送群邀请失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邀请已发送，等待对方确认"})
}

// SearchGroups performs fuzzy search on group names
func (h *GroupHandler) SearchGroups(c *gin.Context) {
	logger.Info.Printf("DEBUG: SearchGroups handler called with path %s", c.Request.URL.Path)
	keyword := c.Query("keyword")
	logger.Info.Printf("DEBUG: keyword = %s", keyword)
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "搜索关键字不能为空"})
		return
	}

	groups, err := h.repo.SearchGroups(keyword)
	if err != nil {
		logger.Info.Printf("DEBUG: SearchGroups error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "搜索群组失败"})
		return
	}

	logger.Info.Printf("DEBUG: Returning %d groups", len(groups))
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

// ApplyJoinGroup creates a group join request
func (h *GroupHandler) ApplyJoinGroup(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		GroupID uint   `json:"group_id" binding:"required"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	// Check if already a member
	isMember, err := h.repo.IsMember(req.GroupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员状态失败"})
		return
	}
	if isMember {
		c.JSON(http.StatusBadRequest, gin.H{"message": "您已是该群组成员"})
		return
	}

	// Check if pending request already exists
	pendingReq, err := h.repo.GetPendingGroupRequest(req.GroupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查入群申请失败"})
		return
	}
	if pendingReq != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "您已有一个待处理的入群申请"})
		return
	}

	// Create request
	groupReq := &model.GroupRequest{
		UserID:  userID,
		GroupID: req.GroupID,
		Remark:  req.Remark,
		Status:  model.GroupRequestStatusPending,
		Type:    model.GroupRequestTypeJoinApply,
	}
	if err := h.repo.CreateGroupRequest(groupReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建入群申请失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "入群申请已提交",
		"data":    groupReq,
	})
}

// GetGroupRequests returns all pending requests for a group
func (h *GroupHandler) GetGroupRequests(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群组ID错误"})
		return
	}

	// Check if current user is group owner or admin
	group, err := h.repo.GetGroupInfo(uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取群组信息失败"})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "群组不存在"})
		return
	}
	operatorRole, err := h.getOperatorRole(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if operatorRole < model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主或管理员可以查看入群申请"})
		return
	}

	requests, err := h.repo.GetGroupRequests(uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取入群申请列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

// HandleGroupRequest accepts or rejects a group join request
func (h *GroupHandler) HandleGroupRequest(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		ReqID  uint   `json:"req_id" binding:"required"`
		Action string `json:"action" binding:"required,oneof=accept reject"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	groupReq, err := h.repo.GetGroupRequestByID(req.ReqID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询入群申请失败"})
		return
	}
	if groupReq == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "入群申请不存在"})
		return
	}
	if groupReq.Type != model.GroupRequestTypeJoinApply {
		c.JSON(http.StatusBadRequest, gin.H{"message": "该请求不是入群申请"})
		return
	}

	// Check if current user is group owner
	group, err := h.repo.GetGroupInfo(groupReq.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取群组信息失败"})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "群组不存在"})
		return
	}

	operatorRole, err := h.getOperatorRole(groupReq.GroupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if operatorRole < model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主或管理员可以处理入群申请"})
		return
	}

	if req.Action == "accept" {
		// Add user as member and update request status
		if err := h.repo.AddMember(groupReq.GroupID, groupReq.UserID, 1); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "添加成员失败"})
			return
		}
		if err := h.repo.UpdateGroupRequestStatus(req.ReqID, model.GroupRequestStatusAccepted); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新申请状态失败"})
			return
		}

		// Add group to conversation list for new member
		if err := h.convRepo.UpsertGroup(groupReq.UserID, groupReq.GroupID, 0, "您已加入群聊", 0); err != nil {
			// Log error but don't fail the request
		}

		c.JSON(http.StatusOK, gin.H{"message": "已接受入群申请"})
	} else if req.Action == "reject" {
		// Update request status to rejected
		if err := h.repo.UpdateGroupRequestStatus(req.ReqID, model.GroupRequestStatusRejected); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新申请状态失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "已拒绝入群申请"})
	}
}

// GetMyInvitations returns pending invitations for current user
func (h *GroupHandler) GetMyInvitations(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	invitations, err := h.repo.GetUserInvitations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取群邀请失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invitations})
}

// HandleGroupInvitation accepts or rejects a group invitation by invitee
func (h *GroupHandler) HandleGroupInvitation(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		ReqID  uint   `json:"req_id" binding:"required"`
		Action string `json:"action" binding:"required,oneof=accept reject"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	invite, err := h.repo.GetGroupRequestByID(req.ReqID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询群邀请失败"})
		return
	}
	if invite == nil || invite.Type != model.GroupRequestTypeInvite {
		c.JSON(http.StatusNotFound, gin.H{"message": "群邀请不存在"})
		return
	}
	if invite.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"message": "您无权处理该邀请"})
		return
	}

	if req.Action == "accept" {
		isMember, err := h.repo.IsMember(invite.GroupID, invite.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员状态失败"})
			return
		}
		if !isMember {
			if err := h.repo.AddMember(invite.GroupID, invite.UserID, 1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "加入群组失败"})
				return
			}
		}

		if err := h.repo.UpdateGroupRequestStatus(req.ReqID, model.GroupRequestStatusAccepted); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新邀请状态失败"})
			return
		}

		if err := h.convRepo.UpsertGroup(invite.UserID, invite.GroupID, 0, "您已加入群聊", 0); err != nil {
			// Log error but don't fail request
		}

		c.JSON(http.StatusOK, gin.H{"message": "已加入群聊"})
		return
	}

	if err := h.repo.UpdateGroupRequestStatus(req.ReqID, model.GroupRequestStatusRejected); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新邀请状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已拒绝群邀请"})
}

// UpdateGroupInfo updates mutable group fields (currently: name).
func (h *GroupHandler) UpdateGroupInfo(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群组ID错误"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" || len([]rune(name)) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群名称长度需在1-50字符"})
		return
	}

	role, err := h.getOperatorRole(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if role < model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主或管理员可以修改群名称"})
		return
	}

	if err := h.repo.UpdateGroupName(uint(groupID), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新群名称失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "群名称已更新"})
}

// UpdateGroupAnnouncement allows owner to publish announcement.
func (h *GroupHandler) UpdateGroupAnnouncement(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群组ID错误"})
		return
	}

	var req struct {
		Announcement string `json:"announcement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	announcement := strings.TrimSpace(req.Announcement)
	if len([]rune(announcement)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群公告长度不能超过500字符"})
		return
	}

	role, err := h.getOperatorRole(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if role != model.GroupRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主可以发布群公告"})
		return
	}

	if err := h.repo.UpdateGroupAnnouncement(uint(groupID), announcement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新群公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "群公告已更新", "data": gin.H{"announcement": announcement}})
}

// UpdateGroupAvatar updates group avatar. Owner/Admin allowed.
func (h *GroupHandler) UpdateGroupAvatar(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群组ID错误"})
		return
	}

	role, err := h.getOperatorRole(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if role < model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主或管理员可以修改群头像"})
		return
	}

	group, err := h.repo.GetGroupInfo(uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取群组信息失败"})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "群组不存在"})
		return
	}

	var req struct {
		Avatar string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	avatar := strings.TrimSpace(req.Avatar)
	if avatar == "" {
		avatar = defaultGroupAvatar(group.Name)
	}
	if len(avatar) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "头像链接过长"})
		return
	}

	if err := h.repo.UpdateGroupAvatar(uint(groupID), avatar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新群头像失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "群头像已更新", "data": gin.H{"avatar": avatar}})
}

// RemoveGroupMember removes one member from group according to role rules.
func (h *GroupHandler) RemoveGroupMember(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		GroupID uint `json:"group_id" binding:"required"`
		UserID  uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	if req.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请使用退群功能，不能删除自己"})
		return
	}

	actorRole, err := h.getOperatorRole(req.GroupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if actorRole < model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主或管理员可以删除成员"})
		return
	}

	targetRole, err := h.getOperatorRole(req.GroupID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查目标成员状态失败"})
		return
	}
	if targetRole == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "目标用户不在群内"})
		return
	}
	if targetRole == model.GroupRoleOwner {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不能删除群主"})
		return
	}
	if actorRole == model.GroupRoleAdmin && targetRole >= model.GroupRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "管理员不能删除管理员"})
		return
	}

	if err := h.repo.RemoveMember(req.GroupID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除成员失败"})
		return
	}
	_ = h.repo.DeleteGroupConversation(req.UserID, req.GroupID)

	c.JSON(http.StatusOK, gin.H{"message": "成员已移除"})
}

// AddGroupAdmin sets an existing member as admin. Owner only.
func (h *GroupHandler) AddGroupAdmin(c *gin.Context) {
	h.changeAdminRole(c, model.GroupRoleAdmin, "已设为管理员")
}

// RemoveGroupAdmin revokes admin role. Owner only.
func (h *GroupHandler) RemoveGroupAdmin(c *gin.Context) {
	h.changeAdminRole(c, model.GroupRoleMember, "已取消管理员")
}

func (h *GroupHandler) changeAdminRole(c *gin.Context, targetRole int8, successMsg string) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		GroupID uint `json:"group_id" binding:"required"`
		UserID  uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	role, err := h.getOperatorRole(req.GroupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if role != model.GroupRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主可以管理管理员"})
		return
	}
	if req.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不能对自己执行该操作"})
		return
	}

	targetCurrentRole, err := h.getOperatorRole(req.GroupID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查目标成员状态失败"})
		return
	}
	if targetCurrentRole == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "目标用户不在群内"})
		return
	}
	if targetCurrentRole == model.GroupRoleOwner {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不能修改群主角色"})
		return
	}

	if err := h.repo.UpdateMemberRole(req.GroupID, req.UserID, targetRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新成员角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": successMsg})
}

// TransferGroupOwnership transfers owner role to another member.
func (h *GroupHandler) TransferGroupOwnership(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		GroupID uint `json:"group_id" binding:"required"`
		UserID  uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}
	if req.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不能转让给自己"})
		return
	}

	role, err := h.getOperatorRole(req.GroupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if role != model.GroupRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主可以转让群主"})
		return
	}

	targetRole, err := h.getOperatorRole(req.GroupID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查目标成员状态失败"})
		return
	}
	if targetRole == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "目标用户不在群内"})
		return
	}

	if err := h.repo.TransferOwnership(req.GroupID, userID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "转让群主失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "群主转让成功"})
}

// DismissGroup dismisses a group. Owner only.
func (h *GroupHandler) DismissGroup(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "群组ID错误"})
		return
	}

	role, err := h.getOperatorRole(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查成员权限失败"})
		return
	}
	if role != model.GroupRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"message": "仅群主可以解散群"})
		return
	}

	if err := h.repo.DismissGroup(uint(groupID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "解散群失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "群已解散"})
}
