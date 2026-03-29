package model

import "time"

// User 用户表
type User struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash  string    `gorm:"type:varchar(255);not null" json:"-"`
	Nickname      string    `gorm:"type:varchar(50)" json:"nickname"`
	AvatarURL     string    `gorm:"type:varchar(255)" json:"avatar_url"`
	Email         string    `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
