package services

import (
	"errors"
	"beam/data/models"
	"beam/data/repositories"
)

type UserService interface {
	Register(user models.User) error
	GetUserByID(uid string) (*models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(uid string) error
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(user models.User) error {
	existingUser, _ := s.userRepo.FindByID(user.UID)
	if existingUser != nil {
		return errors.New("user already exists")
	}
	return s.userRepo.Create(user)
}

func (s *userService) GetUserByID(uid string) (*models.User, error) {
	return s.userRepo.FindByID(uid)
}

func (s *userService) UpdateUser(user models.User) error {
	existingUser, err := s.userRepo.FindByID(user.UID)
	if err != nil || existingUser == nil {
		return errors.New("user not found")
	}
	return s.userRepo.Update(user)
}

func (s *userService) DeleteUser(uid string) error {
	return s.userRepo.Delete(uid)
}
