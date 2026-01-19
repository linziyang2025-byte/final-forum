package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"final-forum/internal/database"
	"final-forum/internal/handlers"
	"final-forum/internal/router"
)

func main() {
	db := database.OpenDB()
	defer db.Close()
	database.ApplySchema(db)

	r := chi.NewRouter()
	r.Use(router.WithCORS)

	r.Post("/auth/register", handlers.Register(db))
	r.Post("/auth/login", handlers.Login(db))
	r.Post("/auth/logout", handlers.Logout)
	r.Get("/auth/me", handlers.AuthMe)
	r.Post("/auth/is-me", handlers.IsMe)

	r.Route("/api", func(apiR chi.Router) {
		apiR.Use(router.RequireLogin)

		apiR.Get("/topics", handlers.ListTopics(db))
		apiR.Post("/topics", handlers.CreateTopic(db))
		apiR.Put("/topics/{topicID}", handlers.UpdateTopic(db))
		apiR.Delete("/topics/{topicID}", handlers.DeleteTopic(db))

		apiR.Get("/topics/{topicID}/posts", handlers.ListPosts(db))
		apiR.Post("/topics/{topicID}/posts", handlers.CreatePost(db))
		apiR.Get("/posts/{postID}", handlers.GetPost(db))
		apiR.Put("/posts/{postID}", handlers.UpdatePost(db))
		apiR.Delete("/posts/{postID}", handlers.DeletePost(db))

		apiR.Get("/posts/{postID}/comments", handlers.ListComments(db))
		apiR.Post("/posts/{postID}/comments", handlers.CreateComment(db))
		apiR.Put("/comments/{commentID}", handlers.UpdateComment(db))
		apiR.Delete("/comments/{commentID}", handlers.DeleteComment(db))
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
