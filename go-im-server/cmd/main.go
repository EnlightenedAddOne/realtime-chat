package main

import (
	"fmt"
	"log"
	"os"

	"go-im-server/config"
	"go-im-server/internal/model"
	"go-im-server/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	// 初始化日志
	logger.Init()
	logger.Info.Println("🚀 启动 IM 服务...")

	// 加载配置
	_, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 连接数据库
	DB, err = gorm.Open(postgres.Open(config.App.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}
	logger.Info.Println("✅ 成功连接到 PostgreSQL")

	// 自动建表
	if err := DB.AutoMigrate(&model.User{}, &model.Message{}, &model.Friend{}, &model.FriendRequest{}, &model.Conversation{}, &model.Group{}, &model.GroupMember{}, &model.GroupRequest{}); err != nil {
		log.Fatalf("❌ 自动建表失败: %v", err)
	}

	if err := migrateConversationPrimaryKey(DB); err != nil {
		log.Fatalf("❌ 会话表主键迁移失败: %v", err)
	}
	logger.Info.Println("✅ 数据库表结构同步完成")

	// 创建上传目录
	if err := os.MkdirAll(config.App.Upload.Path, 0755); err != nil {
		log.Fatalf("❌ 创建上传目录失败: %v", err)
	}

	// 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", config.App.Server.Port)
	logger.Infof("🌐 服务器启动于 http://localhost%s", addr)
	if err := RunServer(addr); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}

func migrateConversationPrimaryKey(db *gorm.DB) error {
	if err := db.Exec(`ALTER TABLE conversations ALTER COLUMN type SET DEFAULT 1`).Error; err != nil {
		return err
	}
	if err := db.Exec(`UPDATE conversations SET type = 1 WHERE type IS NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE conversations ALTER COLUMN type SET NOT NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE conversations DROP CONSTRAINT IF EXISTS conversations_pkey`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE conversations ADD PRIMARY KEY (user_id, peer_id, type)`).Error; err != nil {
		return err
	}
	return nil
}
