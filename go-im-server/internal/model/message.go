package model

import "time"

// Message 消息表
// MsgType: 1-文本, 2-图片, 3-语音, 4-视频
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"index;not null" json:"sender_id"`
	Sender     *User     `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReceiverID *uint     `gorm:"index" json:"receiver_id"`
	GroupID    *uint     `gorm:"index" json:"group_id"`
	MsgType    int8      `gorm:"type:smallint;not null;comment:'1:text, 2:image, 3:voice, 4:video'" json:"msg_type"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Duration   int       `gorm:"type:integer;default:0" json:"duration"`
	IsRead     int8      `gorm:"type:smallint;default:0" json:"is_read"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}
