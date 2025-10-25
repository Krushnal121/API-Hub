package graph

import (
	"context"

	"github.com/Krushnal121/API-Hub/GraphQL/Go/graph/model"
)

// Query Resolvers - Thin layer that delegates to services

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return r.userService.GetAllUsers(ctx)
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	return r.userService.GetUserByID(ctx, id)
}

func (r *queryResolver) Posts(ctx context.Context) ([]*model.Post, error) {
	return r.postService.GetAllPosts(ctx)
}

// Mutation Resolvers - Thin layer that delegates to services

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	return r.userService.CreateUser(ctx, input)
}

func (r *mutationResolver) CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error) {
	return r.postService.CreatePost(ctx, input)
}

func (r *mutationResolver) DeletePost(ctx context.Context, id string) (*model.Post, error) {
	return r.postService.DeletePost(ctx, id)
}

// Auto-generated resolver types (DON'T DELETE)
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
