package api

import (
	"errors"
	"go-im-server/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FriendHandler struct {
	repo *repository.FriendRepository
}

func NewFriendHandler(db *gorm.DB) *FriendHandler {
	return &FriendHandler{repo: repository.NewFriendRepository(db)}
}

// GetFriends 获取好友列表
func (h *FriendHandler) GetFriends(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	friends, err := h.repo.GetFriendList(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取好友失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": friends})
}

// SearchUsers 搜索用户
func (h *FriendHandler) SearchUsers(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请输入关键词"})
		return
	}

	users, err := h.repo.SearchUser(keyword, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "搜索失败"})
		return
	}

	// 补充好友状态 (IsFriend)
	// 这里可以优化，在 SQL 里直接查询，或者循环查询（量小可接受）
	var result []gin.H
	for _, u := range users {
		isFriend, _ := h.repo.CheckIsFriend(userID, u.ID)
		hasPending, _ := h.repo.CheckHasPendingRequest(userID, u.ID)

		result = append(result, gin.H{
			"id":          u.ID,
			"username":    u.Username,
			"nickname":    u.Nickname,
			"avatar_url":  u.AvatarURL,
			"is_friend":   isFriend,
			"has_pending": hasPending,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// SendRequest 发送好友请求
func (h *FriendHandler) SendRequest(c *gin.Context) {
	senderID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		ReceiverID uint   `json:"receiver_id"`
		Remark     string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	// 检查是否已经是好友
	isFriend, _ := h.repo.CheckIsFriend(senderID, req.ReceiverID)
	if isFriend {
		c.JSON(http.StatusBadRequest, gin.H{"message": "已经是好友了"})
		return
	}

	// 检查是否有待处理的请求
	hasPending, _ := h.repo.CheckHasPendingRequest(senderID, req.ReceiverID)
	if hasPending {
		c.JSON(http.StatusBadRequest, gin.H{"message": "已发送过请求，请等待对方处理"})
		return
	}

	if err := h.repo.SendFriendRequest(senderID, req.ReceiverID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送请求失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "发送成功"})
}

// GetPendingRequests 获取待处理请求
func (h *FriendHandler) GetPendingRequests(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	requests, err := h.repo.GetPendingRequests(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取请求列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

// HandleRequest 处理请求
func (h *FriendHandler) HandleRequest(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var body struct {
		ReqID  uint   `json:"req_id"`
		Action string `json:"action"` // accept / reject
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	if err := h.repo.HandleFriendRequest(body.ReqID, userID, body.Action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处理成功"})
}

// DeleteFriend 删除好友
func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req struct {
		FriendID uint `json:"friend_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.FriendID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	if req.FriendID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不能删除自己"})
		return
	}

	if err := h.repo.DeleteFriend(userID, req.FriendID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "对方不是你的好友"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除好友失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除好友"})
}

// Mock AddFriend for testing (User A adds User B directly) - DEPRECATED
func (h *FriendHandler) AddFriend(c *gin.Context) {
	userID, _ := getUserID(c)
	friendIDStr := c.Query("friend_id")
	friendID, _ := strconv.Atoi(friendIDStr)

	// Direct add for testing
	// h.repo.AddFriendship(userID, uint(friendID))
	// c.JSON(200, gin.H{"msg": "added"})

	_ = userID
	_ = friendID
	c.JSON(http.StatusGone, gin.H{"message": "Use SendRequest instead"})
}
