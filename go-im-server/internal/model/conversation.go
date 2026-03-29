package model

import "time"

// Conversation 会话表
type Conversation struct {
	UserID         uint      `gorm:"primaryKey;index" json:"user_id"` // 归属用户
	PeerID         uint      `gorm:"primaryKey;index" json:"peer_id"` // 对方用户ID (或群ID)
	Type           int       `gorm:"primaryKey;default:1;comment:'1:private, 2:group'" json:"type"`
	LastMsgID      uint      `json:"last_msg_id"`
	LastMsgContent string    `gorm:"type:varchar(500)" json:"last_msg_content"` // 消息摘要冗余存储
	UnreadCount    int       `gorm:"default:0" json:"unread_count"`
	UpdatedAt      time.Time `gorm:"index" json:"updated_at"`
	// 非数据库字段，根据 Type 字段动态赋值
	Peer  *User  `gorm:"-" json:"peer,omitempty"`  // Type=1 时存储对方用户信息
	Group *Group `gorm:"-" json:"group,omitempty"` // Type=2 时存储群组信息
}

func (Conversation) TableName() string {
	return "conversations"
}
