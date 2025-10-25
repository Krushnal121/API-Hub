package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Krushnal121/API-Hub/GraphQL/Go/graph/model"
	"github.com/Krushnal121/API-Hub/GraphQL/Go/repository"
)

// UserService handles business logic for users
type UserService interface {
	GetAllUsers(ctx context.Context) ([]*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	CreateUser(ctx context.Context, input model.NewUser) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service with dependency injection
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	// Add business logic here (e.g., filtering, sorting, pagination)
	return s.userRepo.GetAll(ctx)
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	// Add business logic here (e.g., caching, authorization checks)
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	// Business logic: validate email, check duplicates, etc.
	if input.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Generate unique ID (in production, use UUID)
	user := &model.User{
		ID:    fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:  input.Name,
		Email: input.Email,
		Posts: []*model.Post{},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
