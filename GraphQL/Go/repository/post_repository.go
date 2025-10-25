package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/Krushnal121/API-Hub/GraphQL/Go/graph/model"
)

type PostRepository interface {
	GetAll(ctx context.Context) ([]*model.Post, error)
	GetByID(ctx context.Context, id string) (*model.Post, error)
	GetByAuthorID(ctx context.Context, authorID string) ([]*model.Post, error)
	Create(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id string) (*model.Post, error)
}

type InMemoryPostRepository struct {
	posts []*model.Post
	mu    sync.RWMutex
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts: []*model.Post{},
	}
}

func (r *InMemoryPostRepository) GetAll(ctx context.Context) ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]*model.Post, len(r.posts))
	copy(posts, r.posts)
	return posts, nil
}

func (r *InMemoryPostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, post := range r.posts {
		if post.ID == id {
			return post, nil
		}
	}
	return nil, fmt.Errorf("post with id %s not found", id)
}

func (r *InMemoryPostRepository) GetByAuthorID(ctx context.Context, authorID string) ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var authorPosts []*model.Post
	for _, post := range r.posts {
		if post.Author.ID == authorID {
			authorPosts = append(authorPosts, post)
		}
	}
	return authorPosts, nil
}

func (r *InMemoryPostRepository) Create(ctx context.Context, post *model.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.posts = append(r.posts, post)
	return nil
}

func (r *InMemoryPostRepository) Delete(ctx context.Context, id string) (*model.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, post := range r.posts {
		if post.ID == id {
			deleted := post
			r.posts = append(r.posts[:i], r.posts[i+1:]...)
			return deleted, nil
		}
	}
	return nil, fmt.Errorf("post with id %s not found", id)
}
