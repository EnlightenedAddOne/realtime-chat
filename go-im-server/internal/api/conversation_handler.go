package api

import (
	"go-im-server/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConversationHandler struct {
	repo *repository.ConversationRepository
}

func NewConversationHandler(db *gorm.DB) *ConversationHandler {
	return &ConversationHandler{
		repo: repository.NewConversationRepository(db),
	}
}

// GetConversations 获取会话列表
func (h *ConversationHandler) GetConversations(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	conversations, err := h.repo.GetList(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conversations})
}

// MarkRead 标记会话已读
func (h *ConversationHandler) MarkRead(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	type req struct {
		PeerID uint `json:"peer_id"`
		Type   int  `json:"type"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	if r.Type != 1 && r.Type != 2 {
		r.Type = 1
	}

	// 1. 重置未读数
	if err := h.repo.ResetUnread(userID, r.PeerID, r.Type); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败"})
		return
	}

	// 2. (可选) 更新 last_read_msg_id (如果需要的话，目前 conversation 表没用到这个字段，只用 ResetUnread 简单处理)
	// 如果需要精确的 Read Receipt，需要客户端传 msg_id，并更新到 DB

	c.JSON(http.StatusOK, gin.H{"message": "已读"})
}
