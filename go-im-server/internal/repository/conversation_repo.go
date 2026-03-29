package repository

import (
	"go-im-server/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// GetList 获取会话列表，根据 Type 字段分别关联用户和群组信息
func (r *ConversationRepository) GetList(userID uint) ([]model.Conversation, error) {
	var list []model.Conversation

	// 第一步：获取原始会话列表
	err := r.db.Where("user_id = ?", userID).Order("updated_at desc").Find(&list).Error
	if err != nil {
		return nil, err
	}

	// 第二步：收集 userIDs (Type=1) 和 groupIDs (Type=2)
	var userIDs []uint
	var groupIDs []uint

	for _, conv := range list {
		if conv.Type == 1 {
			userIDs = append(userIDs, conv.PeerID)
		} else if conv.Type == 2 {
			groupIDs = append(groupIDs, conv.PeerID)
		}
	}

	// 第三步：批量获取用户
	var users []model.User
	if len(userIDs) > 0 {
		if err := r.db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
	}

	// 第四步：批量获取群组
	var groups []model.Group
	if len(groupIDs) > 0 {
		if err := r.db.Where("id IN ?", groupIDs).Find(&groups).Error; err != nil {
			return nil, err
		}
	}

	// 第五步：构建用户和群组的映射表
	userMap := make(map[uint]*model.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	groupMap := make(map[uint]*model.Group)
	for i := range groups {
		groupMap[groups[i].ID] = &groups[i]
	}

	// 第六步：将用户和群组关联回会话
	for i := range list {
		if list[i].Type == 1 {
			// 私聊：关联用户
			if user, ok := userMap[list[i].PeerID]; ok {
				list[i].Peer = user
			}
		} else if list[i].Type == 2 {
			// 群聊：关联群组
			if group, ok := groupMap[list[i].PeerID]; ok {
				list[i].Group = group
			}
		}
	}

	return list, nil
}

// Upsert 更新或创建会话
// increment: 未读数增量 (发送方为0，接收方为1)
func (r *ConversationRepository) Upsert(userID, peerID uint, lastMsgID uint, lastMsgContent string, increment int) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "peer_id"}, {Name: "type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_msg_id":      lastMsgID,
			"last_msg_content": lastMsgContent,
			"updated_at":       time.Now(),
			"unread_count":     gorm.Expr("conversations.unread_count + ?", increment),
		}),
	}).Create(&model.Conversation{
		UserID:         userID,
		PeerID:         peerID,
		Type:           1, // Private conversation
		LastMsgID:      lastMsgID,
		LastMsgContent: lastMsgContent,
		UnreadCount:    increment,
		UpdatedAt:      time.Now(),
	}).Error
}

// UpsertGroup 更新或创建群组会话
// increment: 未读数增量 (发送方为0，其他成员为1)
// Type 字段值: 1=私聊, 2=群聊
func (r *ConversationRepository) UpsertGroup(userID, groupID uint, lastMsgID uint, lastMsgContent string, increment int) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "peer_id"}, {Name: "type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_msg_id":      lastMsgID,
			"last_msg_content": lastMsgContent,
			"updated_at":       time.Now(),
			"unread_count":     gorm.Expr("conversations.unread_count + ?", increment),
		}),
	}).Create(&model.Conversation{
		UserID:         userID,
		PeerID:         groupID,
		Type:           2, // Group conversation
		LastMsgID:      lastMsgID,
		LastMsgContent: lastMsgContent,
		UnreadCount:    increment,
		UpdatedAt:      time.Now(),
	}).Error
}

// ResetUnread 重置未读数
func (r *ConversationRepository) ResetUnread(userID, peerID uint, convType int) error {
	return r.db.Model(&model.Conversation{}).
		Where("user_id = ? AND peer_id = ? AND type = ?", userID, peerID, convType).
		Update("unread_count", 0).Error
}
