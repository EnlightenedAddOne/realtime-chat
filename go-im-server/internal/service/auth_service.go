package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go-im-server/internal/model"
	"go-im-server/internal/repository"
	"go-im-server/pkg/jwt"
	"go-im-server/pkg/security"
)

var (
	ErrUserExists       = errors.New("用户已存在")
	ErrInvalidLogin     = errors.New("用户名或密码错误")
	ErrInvalidUsername  = errors.New("用户名格式错误")
	ErrInvalidPassword  = errors.New("密码格式错误")
	ErrEmailExists      = errors.New("邮箱已被注册")
	ErrInvalidEmail     = errors.New("邮箱格式错误")
	ErrEmailNotVerified = errors.New("邮箱未验证")
)

// Username validation: 3-20 chars, letters/numbers/underscore, start with letter
var usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,19}$`)

// Email validation
var emailRegex = regexp.MustCompile(`^[\w.-]+@[\w.-]+\.\w+$`)

// Password validation: 6-20 chars, letters and numbers only
var passwordAllowedRegex = regexp.MustCompile(`^[A-Za-z0-9]{6,20}$`)
var nonUsernameCharRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// ValidateUsername checks username format
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

// ValidatePassword checks password format
func ValidatePassword(password string) error {
	if !passwordAllowedRegex.MatchString(password) {
		return ErrInvalidPassword
	}
	return nil
}

// ValidateEmail checks email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func (s *AuthService) Register(username, password string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	existing, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		Nickname:     username,
		AvatarURL:    fmt.Sprintf("https://api.dicebear.com/7.x/micah/svg?seed=%s&backgroundColor=fef3d4", username),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// RegisterWithEmail registers a new user with email and auto-generates username
func (s *AuthService) RegisterWithEmail(password, email string) (*model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	// Check email
	existingEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, ErrEmailExists
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	username, err := s.generateUniqueUsernameByEmail(email)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:      username,
		PasswordHash:  hash,
		Nickname:      username,
		Email:         email,
		EmailVerified: false,
		AvatarURL:     fmt.Sprintf("https://api.dicebear.com/7.x/micah/svg?seed=%s&backgroundColor=fef3d4", username),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, *model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, ErrInvalidLogin
	}

	if !security.CheckPassword(password, user.PasswordHash) {
		return "", nil, ErrInvalidLogin
	}

	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func normalizeEmailLocalPartToUsernameBase(email string) string {
	parts := strings.SplitN(email, "@", 2)
	localPart := parts[0]
	base := nonUsernameCharRegex.ReplaceAllString(localPart, "_")
	base = strings.Trim(base, "_")
	if base == "" {
		base = "user"
	}
	if len(base) > 20 {
		base = base[:20]
	}
	if base[0] < 'A' || (base[0] > 'Z' && base[0] < 'a') || base[0] > 'z' {
		base = "u_" + base
	}
	if len(base) < 3 {
		base = base + strings.Repeat("_", 3-len(base))
	}
	if len(base) > 20 {
		base = base[:20]
	}
	return base
}

func (s *AuthService) generateUniqueUsernameByEmail(email string) (string, error) {
	base := normalizeEmailLocalPartToUsernameBase(email)

	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			suffix := fmt.Sprintf("_%d", i)
			maxBaseLen := 20 - len(suffix)
			trimmedBase := base
			if maxBaseLen < 1 {
				maxBaseLen = 1
			}
			if len(trimmedBase) > maxBaseLen {
				trimmedBase = trimmedBase[:maxBaseLen]
			}
			candidate = trimmedBase + suffix
		}

		existing, err := s.userRepo.GetByUsername(candidate)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return candidate, nil
		}
	}

	return "", ErrUserExists
}
