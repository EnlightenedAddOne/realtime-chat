package ws

import (
	"log"
	"time"

	"go-im-server/internal/model"
	"go-im-server/internal/repository"
)

type Hub struct {
	Clients     map[uint]*Client
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan WSMessage
	OnlineQuery chan OnlineQueryRequest
	msgRepo     *repository.MessageRepository
	convRepo    *repository.ConversationRepository
	groupRepo   *repository.GroupRepository
}

type OnlineQueryRequest struct {
	UserID uint
	Resp   chan bool
}

func (h *Hub) sendToClient(client *Client, msg WSMessage) {
	select {
	case client.Send <- msg:
	default:
		log.Printf("drop ws message type=%s for user=%d: send channel full", msg.Type, client.UserID)
	}
}

func (h *Hub) onlineUserIDs() []uint {
	ids := make([]uint, 0, len(h.Clients))
	for id := range h.Clients {
		ids = append(ids, id)
	}
	return ids
}

func (h *Hub) broadcastPresence(targetUserID uint, isOnline bool, excludeUserID uint) {
	presence := WSMessage{
		Type:         "presence",
		TargetUserID: targetUserID,
		IsOnline:     isOnline,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	for userID, client := range h.Clients {
		if excludeUserID != 0 && userID == excludeUserID {
			continue
		}
		h.sendToClient(client, presence)
	}
}

func NewHub(msgRepo *repository.MessageRepository, convRepo *repository.ConversationRepository, groupRepo *repository.GroupRepository) *Hub {
	return &Hub{
		Clients:     make(map[uint]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan WSMessage, 128),
		OnlineQuery: make(chan OnlineQueryRequest),
		msgRepo:     msgRepo,
		convRepo:    convRepo,
		groupRepo:   groupRepo,
	}
}

func (h *Hub) IsOnline(userID uint) bool {
	resp := make(chan bool, 1)
	h.OnlineQuery <- OnlineQueryRequest{UserID: userID, Resp: resp}
	isOnline := <-resp
	log.Printf("[OnlineQuery] user %d online=%v (total clients=%d)", userID, isOnline, len(h.Clients))
	return isOnline
}

func (h *Hub) Run() {
	for {
		select {
		case query := <-h.OnlineQuery:
			_, ok := h.Clients[query.UserID]
			query.Resp <- ok
		case client := <-h.Register:
			h.Clients[client.UserID] = client
			log.Printf("user %d connected", client.UserID)

			snapshot := WSMessage{
				Type:          "presence_snapshot",
				OnlineUserIDs: h.onlineUserIDs(),
				CreatedAt:     time.Now().Format(time.RFC3339),
			}
			h.sendToClient(client, snapshot)

			h.broadcastPresence(client.UserID, true, client.UserID)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
				log.Printf("user %d disconnected", client.UserID)
				h.broadcastPresence(client.UserID, false, 0)
			}
		case msg := <-h.Broadcast:
			// Handle WebRTC signaling messages (no persistence, direct forward)
			if msg.Type == "signal" {
				if msg.ReceiverID > 0 {
					if receiver, ok := h.Clients[msg.ReceiverID]; ok {
						h.sendToClient(receiver, msg)
					}
				}
				// Signal messages are transient, skip the rest of the broadcast logic
				continue
			}

			// 1. 消息持久化
			if msg.Type == "message" && msg.SenderID > 0 && (msg.ReceiverID > 0 || msg.GroupID > 0) {
				dbMsg := &model.Message{
					SenderID: msg.SenderID,
					ReceiverID: func() *uint {
						if msg.ReceiverID > 0 {
							return &msg.ReceiverID
						} else {
							return nil
						}
					}(),
					GroupID: func() *uint {
						if msg.GroupID > 0 {
							return &msg.GroupID
						} else {
							return nil
						}
					}(),
					MsgType:   msg.MsgType,
					Content:   msg.Content,
					Duration:  msg.Duration,
					IsRead:    0,
					CreatedAt: time.Now(),
				}
				if err := h.msgRepo.Create(dbMsg); err != nil {
					log.Printf("failed to save message: %v", err)
				} else {
					// 回填生成的时间
					msg.CreatedAt = dbMsg.CreatedAt.Format(time.RFC3339)
					msg.ID = dbMsg.ID // 假设 WSMessage 有 ID 字段，如果没有需要加上

					// 生成摘要
					snippet := msg.Content
					if msg.MsgType == 2 {
						snippet = "[图片]"
					} else if msg.MsgType == 3 {
						snippet = "[语音]"
					} else if msg.MsgType == 4 {
						snippet = "[视频]"
					} else if len(snippet) > 100 {
						snippet = snippet[:100] + "..."
					}

					// 更新会话
					if msg.ReceiverID > 0 {
						// 1-to-1 消息
						// Sender: unread_inc = 0
						go h.convRepo.Upsert(msg.SenderID, msg.ReceiverID, dbMsg.ID, snippet, 0)
						// Receiver: unread_inc = 1
						go h.convRepo.Upsert(msg.ReceiverID, msg.SenderID, dbMsg.ID, snippet, 1)
					} else if msg.GroupID > 0 {
						// 群组消息
						memberIDs, err := h.groupRepo.GetMembers(msg.GroupID)
						if err != nil {
							log.Printf("failed to get group members: %v", err)
						} else {
							// 遍历所有群成员更新会话
							for _, memberID := range memberIDs {
								// 发送者: unread_inc = 0, 其他成员: unread_inc = 1
								increment := 1
								if memberID == msg.SenderID {
									increment = 0
								}
								go h.convRepo.UpsertGroup(memberID, msg.GroupID, dbMsg.ID, snippet, increment)
							}
						}
					}
				}
			}

			// 2. 转发给接收者 (1-to-1) 或 群组成员 (1-to-many)
			if msg.ReceiverID > 0 {
				if receiver, ok := h.Clients[msg.ReceiverID]; ok {
					h.sendToClient(receiver, msg)
				}
			} else if msg.GroupID > 0 {
				// 广播给群组所有成员（除发送者外）
				memberIDs, err := h.groupRepo.GetMembers(msg.GroupID)
				if err != nil {
					log.Printf("failed to get group members for broadcast: %v", err)
				} else {
					for _, memberID := range memberIDs {
						// 跳过发送者，只转发给其他成员
						if memberID != msg.SenderID {
							if client, ok := h.Clients[memberID]; ok {
								h.sendToClient(client, msg)
							}
						}
					}
				}
			}

			// 3. (可选) 回显给发送者，确认发送成功并更新时间戳
			/*if sender, ok := h.Clients[msg.SenderID]; ok {
				sender.Send <- msg
			}*/
		}
	}
}
