package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Krushnal121/API-Hub/GraphQL/Go/graph/model"
	"github.com/Krushnal121/API-Hub/GraphQL/Go/repository"
)

type PostService interface {
	GetAllPosts(ctx context.Context) ([]*model.Post, error)
	GetPostsByUser(ctx context.Context, userID string) ([]*model.Post, error)
	CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error)
	DeletePost(ctx context.Context, id string) (*model.Post, error)
}

type postService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) PostService {
	return &postService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *postService) GetAllPosts(ctx context.Context) ([]*model.Post, error) {
	return s.postRepo.GetAll(ctx)
}

func (s *postService) GetPostsByUser(ctx context.Context, userID string) ([]*model.Post, error) {
	return s.postRepo.GetByAuthorID(ctx, userID)
}

func (s *postService) CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error) {
	// Business logic: verify author exists
	author, err := s.userRepo.GetByID(ctx, input.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %w", err)
	}

	// Validate content
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	post := &model.Post{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:   input.Title,
		Content: input.Content,
		Author:  author,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (s *postService) DeletePost(ctx context.Context, id string) (*model.Post, error) {
	// Business logic: check permissions, cascade deletes, etc.
	return s.postRepo.Delete(ctx, id)
}
