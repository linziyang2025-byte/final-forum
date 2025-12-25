package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func listComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		rows, err := db.Query(`SELECT id, post_id, content, author, created_at FROM comments WHERE post_id = ? ORDER BY id ASC`, postID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var out []Comment
		for rows.Next() {
			var c Comment
			if err := rows.Scan(&c.ID, &c.PostID, &c.Content, &c.Author, &c.CreatedAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, c)
		}
		writeJSON(w, 200, out)
	}
}

func createComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}
		user = strings.TrimSpace(user)

		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var in struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		in.Content = strings.TrimSpace(in.Content)
		if in.Content == "" {
			http.Error(w, "content required", 400)
			return
		}

		res, err := db.Exec(`INSERT INTO comments(post_id, content, author) VALUES (?, ?, ?)`, postID, in.Content, user)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id64, _ := res.LastInsertId()
		writeJSON(w, 201, map[string]interface{}{"id": id64})
	}
}

func updateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}

		commentID, err := getID(r, "commentID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM comments WHERE id = ?`, commentID).Scan(&owner); err != nil {
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
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		in.Content = strings.TrimSpace(in.Content)
		if in.Content == "" {
			http.Error(w, "content required", 400)
			return
		}

		if _, err := db.Exec(`UPDATE comments SET content = ? WHERE id = ?`, in.Content, commentID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

func deleteComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}

		commentID, err := getID(r, "commentID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM comments WHERE id = ?`, commentID).Scan(&owner); err != nil {
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

		if _, err := db.Exec(`DELETE FROM comments WHERE id = ?`, commentID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}
