package email

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type EmailService struct {
	smtpHost string
	smtpPort string
	username string
	password string
	fromName string
}

type EmailConfig struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
	FromName string
}

func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		smtpHost: config.SMTPHost,
		smtpPort: config.SMTPPort,
		username: config.Username,
		password: config.Password,
		fromName: config.FromName,
	}
}

// GenerateCode generates a 6-digit verification code
func GenerateCode() string {
	code := make([]byte, 3)
	rand.Read(code)
	num := (int(code[0])*256*256 + int(code[1])*256 + int(code[2])) % 1000000
	return fmt.Sprintf("%06d", num)
}

// SendVerificationCode sends a verification code email
func (s *EmailService) SendVerificationCode(toEmail, code, purpose string) error {
	subject := getSubject(purpose)
	body := getVerificationEmailBody(toEmail, code, purpose)

	return s.Send(toEmail, subject, body)
}

func getSubject(purpose string) string {
	switch purpose {
	case "register":
		return "【实时通讯】邮箱验证码"
	case "login":
		return "【实时通讯】登录验证码"
	case "reset":
		return "【实时通讯】找回密码验证码"
	default:
		return "【实时通讯】验证码"
	}
}

func getVerificationEmailBody(email, code, purpose string) string {
	expireTime := 5 // minutes

	switch purpose {
	case "register":
		return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 30px; }
        .code { font-size: 36px; font-weight: bold; color: #667eea; letter-spacing: 8px; text-align: center; margin: 20px 0; }
        .tips { color: #666666; font-size: 14px; line-height: 1.6; }
        .footer { background-color: #f9f9f9; padding: 20px; text-align: center; color: #999999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>实时通讯</h1>
        </div>
        <div class="content">
            <p>亲爱的用户：</p>
            <p>您好！感谢您注册实时通讯。</p>
            <p>您的注册验证码是：</p>
            <div class="code">%s</div>
            <p class="tips">验证码有效期为 %d 分钟，请尽快完成注册。</p>
            <p class="tips">如果这不是您的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            ---<br>
            实时通讯团队
        </div>
    </div>
</body>
</html>`, code, expireTime)
	case "login":
		return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 30px; }
        .code { font-size: 36px; font-weight: bold; color: #667eea; letter-spacing: 8px; text-align: center; margin: 20px 0; }
        .tips { color: #666666; font-size: 14px; line-height: 1.6; }
        .footer { background-color: #f9f9f9; padding: 20px; text-align: center; color: #999999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>实时通讯</h1>
        </div>
        <div class="content">
            <p>亲爱的用户：</p>
            <p>您好！</p>
            <p>您的登录验证码是：</p>
            <div class="code">%s</div>
            <p class="tips">验证码有效期为 %d 分钟，请尽快完成登录。</p>
            <p class="tips">如果这不是您的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            ---<br>
            实时通讯团队
        </div>
    </div>
</body>
</html>`, code, expireTime)
	default:
		return fmt.Sprintf("您的验证码是：%s，有效期 %d 分钟。", code, expireTime)
	}
}

// Send sends an email with TLS support
func (s *EmailService) Send(to, subject, body string) error {
	// Set content type for HTML email
	contentType := "Content-Type: text/html; charset=UTF-8"

	from := fmt.Sprintf("%s <%s>", s.fromName, s.username)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n\r\n%s",
		from, to, subject, contentType, body)

	// Connect to SMTP server with TLS
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	// Create TLS connection
	tlsConfig := &tls.Config{
		ServerName: s.smtpHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authentication
	var auth smtp.Auth
	if s.username != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.smtpHost)
	}

	if auth != nil {
		err = client.Auth(auth)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Send mail
	err = client.Mail(s.username)
	if err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	err = client.Rcpt(to)
	if err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer wc.Close()

	_, err = wc.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// MockEmailService for development/testing
type MockEmailService struct {
	SentCodes map[string]struct{}
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentCodes: make(map[string]struct{}),
	}
}

func (m *MockEmailService) SendVerificationCode(toEmail, code, purpose string) error {
	m.SentCodes[toEmail+code] = struct{}{}
	fmt.Printf("[MOCK EMAIL] To: %s, Code: %s, Purpose: %s\n", toEmail, code, purpose)
	return nil
}

// RateLimiter for email sending
type RateLimiter struct {
	emailRequests map[string]time.Time
	ipRequests    map[string][]time.Time
	mu            struct{}
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		emailRequests: make(map[string]time.Time),
		ipRequests:    make(map[string][]time.Time),
	}
}

// CanSendToEmail checks if can send to this email (1 minute cooldown)
func (r *RateLimiter) CanSendToEmail(email string) bool {
	email = strings.ToLower(email)
	if lastTime, ok := r.emailRequests[email]; ok {
		if time.Since(lastTime) < time.Minute {
			return false
		}
	}
	r.emailRequests[email] = time.Now()
	return true
}

// CanSendFromIP checks if can send from this IP (5 requests per 5 minutes)
func (r *RateLimiter) CanSendFromIP(ip string) bool {
	now := time.Now()
	requests := r.ipRequests[ip]

	// Filter out old requests
	validRequests := make([]time.Time, 0)
	for _, t := range requests {
		if now.Sub(t) < 5*time.Minute {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= 3 {
		return false
	}

	r.ipRequests[ip] = append(validRequests, now)
	return true
}
