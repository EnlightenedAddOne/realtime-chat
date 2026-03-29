package model

import "time"

// FriendRequest 好友申请表
// Status: 0=待处理, 1=已同意, 2=已拒绝
type FriendRequest struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"index;not null" json:"sender_id"`
	ReceiverID uint      `gorm:"index;not null" json:"receiver_id"`
	Remark     string    `gorm:"type:varchar(255)" json:"remark"`
	Status     int8      `gorm:"type:smallint;default:0;comment:'0:pending, 1:accepted, 2:rejected'" json:"status"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Sender   User `gorm:"foreignKey:SenderID" json:"sender"`
	Receiver User `gorm:"foreignKey:ReceiverID" json:"receiver"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
