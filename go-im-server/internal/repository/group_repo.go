package repository

import (
	"errors"
	"time"

	"go-im-server/internal/model"

	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// CreateGroup creates a new group with initial members in a transaction
func (r *GroupRepository) CreateGroup(group *model.Group, memberIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create group
		if err := tx.Create(group).Error; err != nil {
			return err
		}

		// Add owner as member (role 3)
		ownerMember := &model.GroupMember{
			GroupID:  group.ID,
			UserID:   group.OwnerID,
			Role:     3, // Owner
			JoinedAt: group.CreatedAt,
		}
		if err := tx.Create(ownerMember).Error; err != nil {
			return err
		}

		// Add initial members (role 1: Member)
		for _, userID := range memberIDs {
			if userID == group.OwnerID {
				continue // Skip owner as already added
			}
			member := &model.GroupMember{
				GroupID:  group.ID,
				UserID:   userID,
				Role:     1, // Member
				JoinedAt: group.CreatedAt,
			}
			if err := tx.Create(member).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// AddMember adds a single member to the group
func (r *GroupRepository) AddMember(groupID, userID uint, role int8) error {
	member := &model.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}
	return r.db.Create(member).Error
}

// RemoveMember removes a member from the group
func (r *GroupRepository) RemoveMember(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&model.GroupMember{}).Error
}

// GetMembers returns all user IDs in a group
func (r *GroupRepository) GetMembers(groupID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&model.GroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &userIDs).
		Error
	return userIDs, err
}

// GetUserGroups returns all groups a user belongs to
func (r *GroupRepository) GetUserGroups(userID uint) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).
		Error
	return groups, err
}

// GetGroupInfo retrieves group details
func (r *GroupRepository) GetGroupInfo(groupID uint) (*model.Group, error) {
	var group model.Group
	if err := r.db.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// GetGroupRequestByID gets a group request by id
func (r *GroupRepository) GetGroupRequestByID(reqID uint) (*model.GroupRequest, error) {
	var req model.GroupRequest
	err := r.db.First(&req, reqID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// IsMember checks if a user is a member of a group
func (r *GroupRepository) IsMember(groupID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).
		Error
	return count > 0, err
}

// GetMemberDetails returns full user details for all members in a group
func (r *GroupRepository) GetMemberDetails(groupID uint) ([]model.User, error) {
	var members []model.GroupMemberDetail
	err := r.db.Table("users").
		Select("users.id, users.username, users.nickname, users.avatar_url, group_members.role, group_members.joined_at").
		Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ?", groupID).
		Order("group_members.role DESC, group_members.joined_at ASC").
		Scan(&members).
		Error
	if err != nil {
		return nil, err
	}

	users := make([]model.User, 0, len(members))
	for _, m := range members {
		users = append(users, model.User{ID: m.ID, Username: m.Username, Nickname: m.Nickname, AvatarURL: m.AvatarURL})
	}
	return users, nil
}

// GetMemberDetailsWithRole returns user info with group role.
func (r *GroupRepository) GetMemberDetailsWithRole(groupID uint) ([]model.GroupMemberDetail, error) {
	var members []model.GroupMemberDetail
	err := r.db.Table("users").
		Select("users.id, users.username, users.nickname, users.avatar_url, group_members.role, group_members.joined_at").
		Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ?", groupID).
		Order("group_members.role DESC, group_members.joined_at ASC").
		Scan(&members).
		Error
	return members, err
}

// GetMemberRole returns role of a user in group.
func (r *GroupRepository) GetMemberRole(groupID, userID uint) (int8, bool, error) {
	var member model.GroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return member.Role, true, nil
}

// UpdateMemberRole sets the member role.
func (r *GroupRepository) UpdateMemberRole(groupID, userID uint, role int8) error {
	return r.db.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).Error
}

// UpdateGroupName updates group name.
func (r *GroupRepository) UpdateGroupName(groupID uint, name string) error {
	return r.db.Model(&model.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]interface{}{"name": name, "updated_at": time.Now()}).Error
}

// UpdateGroupAnnouncement updates group announcement and timestamp.
func (r *GroupRepository) UpdateGroupAnnouncement(groupID uint, announcement string) error {
	now := time.Now()
	return r.db.Model(&model.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]interface{}{"announcement": announcement, "announcement_updated_at": now, "updated_at": now}).Error
}

// UpdateGroupAvatar updates group avatar and timestamp.
func (r *GroupRepository) UpdateGroupAvatar(groupID uint, avatar string) error {
	now := time.Now()
	return r.db.Model(&model.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]interface{}{"avatar": avatar, "updated_at": now}).Error
}

// TransferOwnership transfers group owner and adjusts roles in transaction.
func (r *GroupRepository) TransferOwnership(groupID, fromUserID, toUserID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Group{}).
			Where("id = ? AND owner_id = ?", groupID, fromUserID).
			Updates(map[string]interface{}{"owner_id": toUserID, "updated_at": time.Now()}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.GroupMember{}).
			Where("group_id = ? AND user_id = ?", groupID, fromUserID).
			Update("role", model.GroupRoleMember).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.GroupMember{}).
			Where("group_id = ? AND user_id = ?", groupID, toUserID).
			Update("role", model.GroupRoleOwner).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteGroupConversation deletes group conversation entries for one user.
func (r *GroupRepository) DeleteGroupConversation(userID, groupID uint) error {
	return r.db.Where("user_id = ? AND peer_id = ? AND type = ?", userID, groupID, 2).
		Delete(&model.Conversation{}).Error
}

// DismissGroup deletes group and related membership/request/conversation data.
func (r *GroupRepository) DismissGroup(groupID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&model.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", groupID).Delete(&model.GroupRequest{}).Error; err != nil {
			return err
		}
		if err := tx.Where("peer_id = ? AND type = ?", groupID, 2).Delete(&model.Conversation{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Group{}, groupID).Error; err != nil {
			return err
		}
		return nil
	})
}

// SearchGroups performs fuzzy search on group names
func (r *GroupRepository) SearchGroups(keyword string) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.Where("name ILIKE ?", "%"+keyword+"%").
		Find(&groups).
		Error
	return groups, err
}

// CreateGroupRequest creates a new group join request
func (r *GroupRepository) CreateGroupRequest(req *model.GroupRequest) error {
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	return r.db.Create(req).Error
}

// GetGroupRequests returns all pending requests for a group with user preload
func (r *GroupRepository) GetGroupRequests(groupID uint) ([]model.GroupRequest, error) {
	var requests []model.GroupRequest
	err := r.db.Where("group_id = ? AND status = ? AND type = ?", groupID, model.GroupRequestStatusPending, model.GroupRequestTypeJoinApply).
		Preload("User").
		Find(&requests).
		Error
	return requests, err
}

// GetPendingGroupRequest checks if a pending request already exists
func (r *GroupRepository) GetPendingGroupRequest(groupID, userID uint) (*model.GroupRequest, error) {
	var req model.GroupRequest
	err := r.db.Where("group_id = ? AND user_id = ? AND status = ? AND type = ?", groupID, userID, model.GroupRequestStatusPending, model.GroupRequestTypeJoinApply).
		First(&req).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// GetPendingGroupInvite checks if a pending invite already exists
func (r *GroupRepository) GetPendingGroupInvite(groupID, userID uint) (*model.GroupRequest, error) {
	var req model.GroupRequest
	err := r.db.Where("group_id = ? AND user_id = ? AND status = ? AND type = ?", groupID, userID, model.GroupRequestStatusPending, model.GroupRequestTypeInvite).
		First(&req).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// GetUserInvitations returns all pending group invitations for a user
func (r *GroupRepository) GetUserInvitations(userID uint) ([]model.GroupRequest, error) {
	var invitations []model.GroupRequest
	err := r.db.Where("user_id = ? AND status = ? AND type = ?", userID, model.GroupRequestStatusPending, model.GroupRequestTypeInvite).
		Preload("Inviter").
		Preload("Group").
		Find(&invitations).
		Error
	return invitations, err
}

// UpdateGroupRequestStatus updates the status of a group request
func (r *GroupRepository) UpdateGroupRequestStatus(reqID uint, status int) error {
	return r.db.Model(&model.GroupRequest{}).
		Where("id = ?", reqID).
		Update("status", status).
		Update("updated_at", time.Now()).
		Error
}
