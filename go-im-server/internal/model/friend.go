package model

import "time"

// Friend 好友关系表
// Status: 0-申请中, 1-已同意, 2-拉黑
type Friend struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_friend;not null" json:"user_id"`
	FriendID  uint      `gorm:"uniqueIndex:idx_user_friend;not null" json:"friend_id"`
	Status    int8      `gorm:"type:smallint;default:0;comment:'0:pending, 1:accepted, 2:blocked'" json:"status"`
	CreatedAt time.Time `json:"created_at"`

	// 关联: 这里的 FriendUser 对应数据库里的 users 表
	// foreignKey:FriendID 表示用 FriendID 去关联 User 表的 ID
	FriendUser User `gorm:"foreignKey:FriendID" json:"friend_user"`
}

func (Friend) TableName() string {
	return "friends"
}
