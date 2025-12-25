package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	db := mustOpenDB("./forum.db")
	defer db.Close()
	mustApplySchema(db, "./schema.sql")

	r := chi.NewRouter()
	r.Use(withCORS)

	// ---------- Auth (no login required) ----------
	r.Post("/auth/register", register(db))
	r.Post("/auth/login", login(db))
	r.Post("/auth/logout", logout)
	r.Get("/auth/me", authMe)

	// ---------- API (login required for non-GET) ----------
	r.Route("/api", func(api chi.Router) {
		api.Use(requireLogin)

		// topics
		api.Get("/topics", listTopics(db))
		api.Post("/topics", createTopic(db))
		api.Put("/topics/{topicID}", updateTopic(db))
		api.Delete("/topics/{topicID}", deleteTopic(db))

		// posts
		api.Get("/topics/{topicID}/posts", listPosts(db))
		api.Post("/topics/{topicID}/posts", createPost(db))
		api.Get("/posts/{postID}", getPost(db))
		api.Put("/posts/{postID}", updatePost(db))
		api.Delete("/posts/{postID}", deletePost(db))

		// comments
		api.Get("/posts/{postID}/comments", listComments(db))
		api.Post("/posts/{postID}/comments", createComment(db))
		api.Put("/comments/{commentID}", updateComment(db))
		api.Delete("/comments/{commentID}", deleteComment(db))
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
