package repository

import (
	"errors"
	"go-im-server/internal/model"

	"gorm.io/gorm"
)

type FriendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) *FriendRepository {
	return &FriendRepository{db: db}
}

// GetFriendList 获取当前用户的好友列表 (包含 User 详情)
func (r *FriendRepository) GetFriendList(userID uint) ([]model.Friend, error) {
	var friends []model.Friend
	// Preload 关联的 User 信息
	// 假设 Friend 模型里定义了 FriendUser 字段
	err := r.db.Where("user_id = ?", userID).Preload("FriendUser").Find(&friends).Error
	return friends, err
}

// SearchUser 模糊搜索用户 (排除自己)
func (r *FriendRepository) SearchUser(keyword string, excludeID uint) ([]model.User, error) {
	var users []model.User
	// ILIKE 用于 PostgreSQL 不区分大小写
	// 注意 SQL 注入，Gorm 使用 ? 绑定参数是安全的
	likeKeyword := "%" + keyword + "%"
	err := r.db.Where("(username ILIKE ? OR nickname ILIKE ?) AND id != ?", likeKeyword, likeKeyword, excludeID).
		Limit(20).
		Find(&users).Error
	return users, err
}

// CheckIsFriend 检查是否已是好友
func (r *FriendRepository) CheckIsFriend(userID, friendID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error
	return count > 0, err
}

// CheckHasPendingRequest 检查是否有待处理的请求
func (r *FriendRepository) CheckHasPendingRequest(senderID, receiverID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.FriendRequest{}).
		Where("sender_id = ? AND receiver_id = ? AND status = 0", senderID, receiverID).
		Count(&count).Error
	return count > 0, err
}

// SendFriendRequest 发送好友请求
func (r *FriendRepository) SendFriendRequest(senderID, receiverID uint, remark string) error {
	req := model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Remark:     remark,
		Status:     0, // Pending
	}
	return r.db.Create(&req).Error
}

// GetPendingRequests 获取当前用户收到的待处理请求
func (r *FriendRepository) GetPendingRequests(receiverID uint) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := r.db.Where("receiver_id = ? AND status = 0", receiverID).
		Preload("Sender"). // 预加载发送者信息
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// HandleFriendRequest 处理好友请求 (同意/拒绝)
func (r *FriendRepository) HandleFriendRequest(reqID, userID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var req model.FriendRequest
		if err := tx.First(&req, reqID).Error; err != nil {
			return err
		}

		// 权限检查
		if req.ReceiverID != userID {
			return errors.New("无权处理该请求")
		}
		if req.Status != 0 {
			return errors.New("请求已被处理")
		}

		status := int8(2) // 默认拒绝
		if action == "accept" {
			status = 1
			// 同意 -> 双方建立好友关系
			// A -> B
			f1 := model.Friend{UserID: req.SenderID, FriendID: req.ReceiverID}
			if err := tx.Create(&f1).Error; err != nil {
				return err
			}
			// B -> A
			f2 := model.Friend{UserID: req.ReceiverID, FriendID: req.SenderID}
			if err := tx.Create(&f2).Error; err != nil {
				return err
			}
		}

		// 更新请求状态
		req.Status = status
		return tx.Save(&req).Error
	})
}

// DeleteFriend 删除双向好友关系，并清理双方之间的待处理好友请求
func (r *FriendRepository) DeleteFriend(userID, friendID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID).
			Delete(&model.Friend{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, friendID, friendID, userID).
			Delete(&model.FriendRequest{}).Error; err != nil {
			return err
		}

		return nil
	})
}
