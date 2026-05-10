package services

import (
	"dorm-man/internal/models"
	"dorm-man/internal/repositories"
)

type UserService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.userRepository.Create(user)
}
