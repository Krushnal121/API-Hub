package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Krushnal121/API-Hub/GraphQL/Go/graph"
	"github.com/Krushnal121/API-Hub/GraphQL/Go/repository"
	"github.com/Krushnal121/API-Hub/GraphQL/Go/service"
)

const defaultPort = "8080"

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = defaultPort
    }

    // Initialize repositories (data layer)
    userRepo := repository.NewInMemoryUserRepository()
    postRepo := repository.NewInMemoryPostRepository()

    // Initialize services (business logic layer)
    userService := service.NewUserService(userRepo)
    postService := service.NewPostService(postRepo, userRepo)

    // Initialize resolver with dependency injection
    resolver := graph.NewResolver(userService, postService)

    // Create GraphQL server
    srv := handler.NewDefaultServer(
        graph.NewExecutableSchema(graph.Config{Resolvers: resolver}),
    )

    // Setup routes
    http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
    http.Handle("/query", srv)

    log.Printf("Connect to http://localhost:%s/ for GraphQL playground", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}