package api

import (
	"net/http"
	"strings"

	"go-im-server/internal/ws"
	"go-im-server/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	Hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{Hub: hub}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *WSHandler) Handle(c *gin.Context) {
	// 1. 获取 token (query 或 header)
	token := c.Query("token")
	if token == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "缺少 token"})
		return
	}

	claims, err := jwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "token 无效"})
		return
	}

	// 2. 升级为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan ws.WSMessage, 64),
		Hub:    h.Hub,
	}

	h.Hub.Register <- client

	// 3. 读写协程
	go client.WritePump()
	client.ReadPump()
}
