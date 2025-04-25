package services

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	
	"black-lotus/internal/models"
	"black-lotus/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
    // Check if user already exists with this email
    existingUser, err := s.userRepo.GetUserByEmail(ctx, input.Email)
    if err != nil {
        return nil, err
    }
    
    if existingUser != nil {
        return nil, errors.New("user with this email already exists")
    }
    
    // Hash password if provided
    var hashedPassword *string
    if input.Password != nil {
        hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        
        hashStr := string(hash)
        hashedPassword = &hashStr
    }
    
    // Create user
    user, err := s.userRepo.CreateUser(ctx, input, hashedPassword)
    if err != nil {
        return nil, err
    }
    
    // Remove sensitive data before returning
    user.HashedPassword = nil
    
    return user, nil
}