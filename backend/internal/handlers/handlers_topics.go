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

func ListTopics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, name, author, created_at FROM topics ORDER BY id DESC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var out []models.Topic
		for rows.Next() {
			var t models.Topic
			if err := rows.Scan(&t.ID, &t.Name, &t.Author, &t.CreatedAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, t)
		}
		api.WriteJSON(w, 200, out)
	}
}

func CreateTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := router.MustUser(w, r)
		if !ok {
			return
		}
		user = strings.TrimSpace(user)

		var in struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" {
			http.Error(w, "name required", 400)
			return
		}

		res, err := db.Exec(`INSERT INTO topics(name, author) VALUES ($1, $2)`, in.Name, user)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id64, _ := res.LastInsertId()
		api.WriteJSON(w, 201, map[string]interface{}{"id": id64})
	}
}

func UpdateTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := router.MustUser(w, r)
		if !ok {
			return
		}

		id, err := api.GetID(r, "topicID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM topics WHERE id = $1`, id).Scan(&owner); err != nil {
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
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" {
			http.Error(w, "name required", 400)
			return
		}

		if _, err := db.Exec(`UPDATE topics SET name = $1 WHERE id = $2`, in.Name, id); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

func DeleteTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := router.MustUser(w, r)
		if !ok {
			return
		}

		id, err := api.GetID(r, "topicID")
		if err != nil {
			http.Error(w, "invalid id", 400)
			return
		}

		var owner string
		if err := db.QueryRow(`SELECT author FROM topics WHERE id = $1`, id).Scan(&owner); err != nil {
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

		if _, err := db.Exec(`DELETE FROM topics WHERE id = $1`, id); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}
