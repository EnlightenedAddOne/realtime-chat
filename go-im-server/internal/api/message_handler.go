package api

import (
	"go-im-server/internal/model"
	"go-im-server/internal/repository"
	"go-im-server/pkg/jwt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageHandler struct {
	msgRepo *repository.MessageRepository
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	repo := repository.NewMessageRepository(db)
	return &MessageHandler{msgRepo: repo}
}

// GetHistory 获取历史消息
func (h *MessageHandler) GetHistory(c *gin.Context) {
	// 1. 获取当前用户 ID
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	// 2. 解析参数
	offsetStr := c.DefaultQuery("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)
	limit := 50

	var messages []model.Message
	var queryErr error

	groupIDStr := c.Query("group_id")
	if groupIDStr != "" {
		groupID, err := strconv.Atoi(groupIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "group_id 错误"})
			return
		}
		messages, queryErr = h.msgRepo.GetHistory(userID, uint(groupID), true, offset, limit)
	} else {
		friendIDStr := c.Query("friend_id")
		friendID, err := strconv.Atoi(friendIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "friend_id 错误"})
			return
		}
		messages, queryErr = h.msgRepo.GetHistory(userID, uint(friendID), false, offset, limit)
	}

	if queryErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败"})
		return
	}

	// 反转顺序：从旧到新返回给前端（因为查询是 ORDER BY DESC）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// 辅助函数：从 header 获取 userID
func getUserID(c *gin.Context) (uint, error) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return 0, http.ErrNoCookie
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	claims, err := jwt.ParseToken(token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
