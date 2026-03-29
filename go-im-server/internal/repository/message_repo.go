package repository

import (
	"go-im-server/internal/model"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *MessageRepository) GetHistory(userID, peerID uint, isGroup bool, offset, limit int) ([]model.Message, error) {
	var messages []model.Message
	var err error

	if isGroup {
		err = r.db.Where("group_id = ?", peerID).
			Preload("Sender").Order("created_at DESC").Offset(offset).Limit(limit).Find(&messages).Error
	} else {
		err = r.db.Where(
			"((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND group_id IS NULL",
			userID, peerID, peerID, userID,
		).Preload("Sender").Order("created_at DESC").Offset(offset).Limit(limit).Find(&messages).Error
	}

	return messages, err
}
