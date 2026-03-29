package api

import (
	"go-im-server/internal/repository"
	"go-im-server/internal/service"
	"go-im-server/internal/ws"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *service.UserService
	hub         *ws.Hub
}

func NewUserHandler(db *gorm.DB, hub *ws.Hub) *UserHandler {
	userRepo := repository.NewUserRepository(db)
	return &UserHandler{userService: service.NewUserService(userRepo), hub: hub}
}

type updateProfileRequest struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Reuse getUserID from message_handler.go (same package)
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.Nickname, req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"user":    user,
	})
}

func (h *UserHandler) GetOnlineStatus(c *gin.Context) {
	if _, err := getUserID(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少 user_id"})
		return
	}

	targetID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user_id 错误"})
		return
	}

	isOnline := false
	if h.hub != nil {
		isOnline = h.hub.IsOnline(uint(targetID))
	} else {
		log.Printf("[GetOnlineStatus] hub is nil for user %d", targetID)
	}

	log.Printf("[GetOnlineStatus] responding is_online=%v for user %d", isOnline, targetID)
	c.JSON(http.StatusOK, gin.H{"is_online": isOnline})
}
