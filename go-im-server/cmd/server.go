package main

import (
	"go-im-server/config"
	"go-im-server/internal/api"
	"go-im-server/internal/repository"
	"go-im-server/internal/ws"
	"go-im-server/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RunServer(addr string) error {
	gin.SetMode(config.App.Server.Mode)
	r := gin.Default()
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		return err
	}

	// 启用 CORS 中间件
	r.Use(CORSMiddleware())

	// 初始化 Repo
	msgRepo := repository.NewMessageRepository(DB)
	convRepo := repository.NewConversationRepository(DB)
	groupRepo := repository.NewGroupRepository(DB)

	// WebSocket Hub (注入 Repo)
	hub := ws.NewHub(msgRepo, convRepo, groupRepo)
	go hub.Run()

	// 静态文件服务 (上传的文件)
	r.Static("/uploads", config.App.Upload.Path)

	// API 路由
	setupRoutes(r, DB, hub)

	return r.Run(addr)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func setupRoutes(r *gin.Engine, db *gorm.DB, hub *ws.Hub) {

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 认证接口
	authHandler := api.NewAuthHandler(db)
	r.POST("/api/auth/send-code", authHandler.SendVerificationCode)
	r.POST("/api/auth/verify-code", authHandler.VerifyCode)
	r.POST("/api/register", authHandler.Register)
	r.POST("/api/login", authHandler.Login)

	// 业务接口 (GET 历史记录 & 好友 & 上传)
	msgHandler := api.NewMessageHandler(db)
	r.GET("/api/messages", msgHandler.GetHistory)

	friendHandler := api.NewFriendHandler(db)
	r.GET("/api/friends", friendHandler.GetFriends)
	r.GET("/api/friends/search", friendHandler.SearchUsers)
	r.POST("/api/friends/request", friendHandler.SendRequest)
	r.GET("/api/friends/requests", friendHandler.GetPendingRequests)
	r.POST("/api/friends/handle", friendHandler.HandleRequest)
	r.POST("/api/friends/delete", friendHandler.DeleteFriend)

	userHandler := api.NewUserHandler(db, hub)
	r.GET("/api/user/profile", userHandler.GetProfile)
	r.POST("/api/user/profile", userHandler.UpdateProfile)
	r.GET("/api/user/online", userHandler.GetOnlineStatus)

	convHandler := api.NewConversationHandler(db)
	r.GET("/api/conversations", convHandler.GetConversations)
	r.POST("/api/conversations/read", convHandler.MarkRead)

	uploadHandler := api.NewUploadHandler()
	r.POST("/api/upload", uploadHandler.UploadFile)

	// 群组接口
	groupHandler := api.NewGroupHandler(db)
	// NOTE: Search route moved to /api/search/groups to avoid route conflicts
	logger.Info.Printf("DEBUG: Registering /api/search/groups route")
	r.GET("/api/search/groups", groupHandler.SearchGroups)

	logger.Info.Printf("DEBUG: Registering /api/groups route (POST)")
	r.POST("/api/groups", groupHandler.CreateGroup)
	logger.Info.Printf("DEBUG: Registering /api/groups/:id/members route")
	r.GET("/api/groups/:id/members", groupHandler.GetGroupMembers)
	logger.Info.Printf("DEBUG: Registering /api/groups/join route")
	r.POST("/api/groups/join", groupHandler.JoinGroup)
	logger.Info.Printf("DEBUG: Registering /api/groups/apply route")
	r.POST("/api/groups/apply", groupHandler.ApplyJoinGroup)
	logger.Info.Printf("DEBUG: Registering /api/groups/:id/requests route")
	r.GET("/api/groups/:id/requests", groupHandler.GetGroupRequests)
	logger.Info.Printf("DEBUG: Registering /api/groups/requests/handle route")
	r.POST("/api/groups/requests/handle", groupHandler.HandleGroupRequest)
	logger.Info.Printf("DEBUG: Registering /api/groups/invitations route")
	r.GET("/api/groups/invitations", groupHandler.GetMyInvitations)
	logger.Info.Printf("DEBUG: Registering /api/groups/invitations/handle route")
	r.POST("/api/groups/invitations/handle", groupHandler.HandleGroupInvitation)
	logger.Info.Printf("DEBUG: Registering /api/groups/:id/info route")
	r.PATCH("/api/groups/:id/info", groupHandler.UpdateGroupInfo)
	logger.Info.Printf("DEBUG: Registering /api/groups/:id/announcement route")
	r.PATCH("/api/groups/:id/announcement", groupHandler.UpdateGroupAnnouncement)
	logger.Info.Printf("DEBUG: Registering /api/groups/:id/avatar route")
	r.PATCH("/api/groups/:id/avatar", groupHandler.UpdateGroupAvatar)
	logger.Info.Printf("DEBUG: Registering /api/groups/members/remove route")
	r.POST("/api/groups/members/remove", groupHandler.RemoveGroupMember)
	logger.Info.Printf("DEBUG: Registering /api/groups/admins/add route")
	r.POST("/api/groups/admins/add", groupHandler.AddGroupAdmin)
	logger.Info.Printf("DEBUG: Registering /api/groups/admins/remove route")
	r.POST("/api/groups/admins/remove", groupHandler.RemoveGroupAdmin)
	logger.Info.Printf("DEBUG: Registering /api/groups/transfer route")
	r.POST("/api/groups/transfer", groupHandler.TransferGroupOwnership)
	logger.Info.Printf("DEBUG: Registering /api/groups/:id/dismiss route")
	r.POST("/api/groups/:id/dismiss", groupHandler.DismissGroup)

	// WebSocket
	wsHandler := api.NewWSHandler(hub)
	r.GET("/ws", wsHandler.Handle)

	// TODO: 添加更多路由
	logger.Info.Println("📝 路由已注册")
}
