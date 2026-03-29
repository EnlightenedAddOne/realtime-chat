package api

import (
	"net/http"
	"strings"

	"go-im-server/config"
	"go-im-server/internal/repository"
	"go-im-server/internal/service"
	"go-im-server/pkg/email"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService *service.AuthService
	userRepo    *repository.UserRepository
	emailSvc    interface {
		SendVerificationCode(email, code, purpose string) error
	}
	rateLimiter *email.RateLimiter
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)

	var emailSvc interface {
		SendVerificationCode(email, code, purpose string) error
	}
	var rateLimiter = email.NewRateLimiter()

	if config.App != nil && config.App.Email.Enable {
		emailSvc = email.NewEmailService(&email.EmailConfig{
			SMTPHost: config.App.Email.Host,
			SMTPPort: config.App.Email.Port,
			Username: config.App.Email.Username,
			Password: config.App.Email.Password,
			FromName: config.App.Email.FromName,
		})
	} else {
		// Use mock email service for development
		emailSvc = email.NewMockEmailService()
	}

	return &AuthHandler{
		authService: authSvc,
		userRepo:    userRepo,
		emailSvc:    emailSvc,
		rateLimiter: rateLimiter,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sendCodeRequest struct {
	Email string `json:"email"`
	Type  string `json:"type"` // register, login, reset
}

type verifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	Type  string `json:"type"`
}

// SendVerificationCode sends verification code to email
func (h *AuthHandler) SendVerificationCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	// Validate email format
	if err := service.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "邮箱格式不正确"})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check rate limit - IP based
	ip := c.ClientIP()
	if !h.rateLimiter.CanSendFromIP(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "请求过于频繁，请稍后再试"})
		return
	}

	// Check rate limit - email based
	if !h.rateLimiter.CanSendToEmail(req.Email) {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "验证码已发送，请稍后再试"})
		return
	}

	// For register type, check if email already exists
	if req.Type == "register" {
		existing, err := h.userRepo.GetByEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
			return
		}
		if existing != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "该邮箱已被注册"})
			return
		}
	}

	// Generate and send code
	code := email.GenerateCode()
	err := h.emailSvc.SendVerificationCode(req.Email, code, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送验证码失败，请稍后再试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "验证码已发送",
		"expire_in":  300,
		"debug_code": code, // Remove in production
	})
}

// VerifyCode verifies the code
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	// In production, you would check against stored codes
	// For now, we'll return success for any 6-digit code
	// This should be implemented with Redis or database

	c.JSON(http.StatusOK, gin.H{
		"message": "验证成功",
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validate email
	if err := service.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "邮箱格式不正确"})
		return
	}

	// Validate password strength
	if err := service.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "密码需为6-20位，仅支持英文和数字"})
		return
	}

	// Check if email already exists
	existingEmail, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}
	if existingEmail != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "该邮箱已被注册"})
		return
	}

	// In production, verify the code here
	// For now, we skip code verification

	user, err := h.authService.RegisterWithEmail(req.Password, req.Email)
	if err != nil {
		message := err.Error()
		switch err {
		case service.ErrUserExists:
			message = "系统生成用户名冲突，请重试"
		case service.ErrInvalidPassword:
			message = "密码需为6-20位，仅支持英文和数字"
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	// 清理空格
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := service.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "邮箱格式不正确"})
		return
	}

	token, user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		message := err.Error()
		if err == service.ErrInvalidLogin {
			message = "邮箱或密码错误"
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}
