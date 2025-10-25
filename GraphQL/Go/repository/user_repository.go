package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/Krushnal121/API-Hub/GraphQL/Go/graph/model"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetAll(ctx context.Context) ([]*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}

// InMemoryUserRepository is a fake repository for demonstration
// In production, this would be a PostgresUserRepository, MongoUserRepository, etc.
type InMemoryUserRepository struct {
	users []*model.User
	mu    sync.RWMutex // Thread-safe for concurrent GraphQL resolvers
}

// NewInMemoryUserRepository creates a new repository with sample data
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: []*model.User{
			{
				ID:    "1",
				Name:  "Alice Johnson",
				Email: "alice@example.com",
				Posts: []*model.Post{},
			},
			{
				ID:    "2",
				Name:  "Bob Smith",
				Email: "bob@example.com",
				Posts: []*model.Post{},
			},
		},
	}
}

func (r *InMemoryUserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	users := make([]*model.User, len(r.users))
	copy(users, r.users)
	return users, nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user with id %s not found", id)
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate IDs
	for _, u := range r.users {
		if u.ID == user.ID {
			return fmt.Errorf("user with id %s already exists", user.ID)
		}
	}

	r.users = append(r.users, user)
	return nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, user := range r.users {
		if user.ID == id {
			r.users = append(r.users[:i], r.users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("user with id %s not found", id)
}
