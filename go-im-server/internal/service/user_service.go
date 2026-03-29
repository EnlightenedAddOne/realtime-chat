package service

import (
	"errors"
	"go-im-server/internal/model"
	"go-im-server/internal/repository"
)

var (
	ErrUserNotFound = errors.New("用户不存在")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUser(uid uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateProfile(uid uint, nickname, avatar string) (*model.User, error) {
	user, err := s.userRepo.GetByID(uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if nickname != "" {
		user.Nickname = nickname
	}
	if avatar != "" {
		user.AvatarURL = avatar
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}
