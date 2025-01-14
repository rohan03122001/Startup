package service

import (
	"errors"

	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService{
	return &UserService{
		userRepo: userRepo,
	}
}


func(s *UserService)CreateUser(username string)(*models.User, error){
	//Validate UserName
	if(len(username) < 3 || len(username) > 20){
		return nil, errors.New("username should be between 3 to 20 characters")
	}

	//check if already exists
	existing, err := s.userRepo.GetByUsername(username)
	if err == nil && existing !=nil{
		return nil, errors.New("username already exist")
	}

	user := &models.User{
		Username: username,
	}

	if err := s.userRepo.CreateUser(user); err != nil{
		return nil, err
	}

	return user, nil
}


func(s *UserService)GetUserProfile(userID string)(*models.User, error){
	return s.userRepo.GetByID(userID)
}

// ValidateUsername checks if username is available
func (s *UserService) ValidateUsername(username string) error {
    if len(username) < 3 || len(username) > 20 {
        return errors.New("username must be between 3 and 20 characters")
    }

    existing, _ := s.userRepo.GetByUsername(username)
    if existing != nil {
        return errors.New("username already taken")
    }

    return nil
}