package graph

import "github.com/Krushnal121/API-Hub/GraphQL/Go/service"

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
// It serves as dependency injection for your app.
type Resolver struct {
	userService service.UserService
	postService service.PostService
}

// NewResolver creates a new resolver with injected dependencies
func NewResolver(userService service.UserService, postService service.PostService) *Resolver {
	return &Resolver{
		userService: userService,
		postService: postService,
	}
}
