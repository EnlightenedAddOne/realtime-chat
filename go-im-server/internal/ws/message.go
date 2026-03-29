package ws

import "time"

// WSMessage 前后端 WebSocket 消息格式
type WSMessage struct {
	ID            uint   `json:"id,omitempty"`
	Type          string `json:"type"` // message/system/signal/presence/presence_snapshot
	SenderID      uint   `json:"sender_id"`
	ReceiverID    uint   `json:"receiver_id"`
	GroupID       uint   `json:"group_id"`
	MsgType       int8   `json:"msg_type"` // 1=text, 2=image, 3=voice, 4=video
	Content       string `json:"content"`
	Duration      int    `json:"duration"`
	CreatedAt     string `json:"created_at"`
	TargetUserID  uint   `json:"target_user_id,omitempty"`
	IsOnline      bool   `json:"is_online,omitempty"`
	OnlineUserIDs []uint `json:"online_user_ids,omitempty"`
}

func NewSystemMessage(content string) WSMessage {
	return WSMessage{
		Type:      "system",
		Content:   content,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}
