package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"final-forum/internal/api"
	"final-forum/internal/models"
	"final-forum/internal/router"
)

func ListPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicID, err := api.GetID(r, "topicID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		rows, err := db.Query(`SELECT id, topic_id, title, content, author, created_at FROM posts WHERE topic_id = $1 ORDER BY id DESC`, topicID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var out []models.Post
		for rows.Next() {
			var p models.Post
			if err := rows.Scan(&p.ID, &p.TopicID, &p.Title, &p.Content, &p.Author, &p.CreatedAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, p)
		}
		api.WriteJSON(w, 200, out)
	}
}

func CreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := router.MustUser(w, r)
		if !ok {
			return
		}
		user = strings.TrimSpace(user)

		topicID, err := api.GetID(r, "topicID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var in struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		in.Title = strings.TrimSpace(in.Title)
		in.Content = strings.TrimSpace(in.Content)
		if in.Title == "" {
			http.Error(w, "title required", 400)
			return
		}

		res, err := db.Exec(`INSERT INTO posts(topic_id, title, content, author) VALUES ($1, $2, $3, $4)`, topicID, in.Title, in.Content, user)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id64, _ := res.LastInsertId()
		api.WriteJSON(w, 201, map[string]interface{}{"id": id64})
	}
}

func GetPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := api.GetID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var p models.Post
		err = db.QueryRow(`SELECT id, topic_id, title, content, author, created_at FROM posts WHERE id = $1`, postID).
			Scan(&p.ID, &p.TopicID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
		if err != nil {
			if api.ErrorsIs(err, sql.ErrNoRows) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		api.WriteJSON(w, 200, p)
	}
}

func UpdatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := router.MustUser(w, r)
		if !ok {
			return
		}

		postID, err := api.GetID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM posts WHERE id = $1`, postID).Scan(&owner); err != nil {
			if api.ErrorsIs(err, sql.ErrNoRows) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		if strings.TrimSpace(owner) != strings.TrimSpace(user) {
			http.Error(w, "forbidden", 403)
			return
		}

		var in struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		in.Title = strings.TrimSpace(in.Title)
		in.Content = strings.TrimSpace(in.Content)
		if in.Title == "" {
			http.Error(w, "title required", 400)
			return
		}

		if _, err := db.Exec(`UPDATE posts SET title = $1, content = $2 WHERE id = $3`, in.Title, in.Content, postID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

func DeletePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := router.MustUser(w, r)
		if !ok {
			return
		}

		postID, err := api.GetID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM posts WHERE id = $1`, postID).Scan(&owner); err != nil {
			if api.ErrorsIs(err, sql.ErrNoRows) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		if strings.TrimSpace(owner) != strings.TrimSpace(user) {
			http.Error(w, "forbidden", 403)
			return
		}

		if _, err := db.Exec(`DELETE FROM posts WHERE id = $1`, postID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}
