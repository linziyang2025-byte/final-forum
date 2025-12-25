package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func listPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		rows, err := db.Query(`SELECT id, topic_id, title, content, author, created_at FROM posts WHERE topic_id = ? ORDER BY id DESC`, topicID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var out []Post
		for rows.Next() {
			var p Post
			if err := rows.Scan(&p.ID, &p.TopicID, &p.Title, &p.Content, &p.Author, &p.CreatedAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, p)
		}
		writeJSON(w, 200, out)
	}
}

func createPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}
		user = strings.TrimSpace(user)

		topicID, err := getID(r, "topicID")
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

		res, err := db.Exec(`INSERT INTO posts(topic_id, title, content, author) VALUES (?, ?, ?, ?)`, topicID, in.Title, in.Content, user)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id64, _ := res.LastInsertId()
		writeJSON(w, 201, map[string]interface{}{"id": id64})
	}
}

func getPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var p Post
		err = db.QueryRow(`SELECT id, topic_id, title, content, author, created_at FROM posts WHERE id = ?`, postID).
			Scan(&p.ID, &p.TopicID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
		if err != nil {
			if errorsIs(err, sql.ErrNoRows) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, 200, p)
	}
}

func updatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}

		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM posts WHERE id = ?`, postID).Scan(&owner); err != nil {
			if errorsIs(err, sql.ErrNoRows) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		if owner != user {
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

		if _, err := db.Exec(`UPDATE posts SET title = ?, content = ? WHERE id = ?`, in.Title, in.Content, postID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

func deletePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}

		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM posts WHERE id = ?`, postID).Scan(&owner); err != nil {
			if errorsIs(err, sql.ErrNoRows) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		if owner != user {
			http.Error(w, "forbidden", 403)
			return
		}

		if _, err := db.Exec(`DELETE FROM posts WHERE id = ?`, postID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}
