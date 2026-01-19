package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"final-forum/internal/api"
	"final-forum/internal/models"
	"final-forum/internal/router"

	"golang.org/x/crypto/bcrypt"
)

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.AuthInfo
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		in.Username = strings.TrimSpace(in.Username)
		if in.Username == "" || in.Password == "" {
			http.Error(w, "username and password required", 400)
			return
		}
		var exists int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM users WHERE username = $1",
			in.Username,
		).Scan(&exists)
		if err != nil {
			http.Error(w, "db error", 500)
			return
		}
		if exists > 0 {
			http.Error(w, "username already exists", 400)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "hash error", 500)
			return
		}

		_, err = db.Exec(
			"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
			in.Username,
			string(hash),
		)
		if err != nil {
			http.Error(w, "insert failed", 500)
			return
		}

		api.WriteJSON(w, 201, map[string]string{
			"message": "registered",
		})
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.AuthInfo
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		in.Username = strings.TrimSpace(in.Username)
		if in.Username == "" || in.Password == "" {
			http.Error(w, "invalid username or password", 400)
			return
		}

		var hash string
		if err := db.QueryRow(`SELECT password_hash FROM users WHERE username = $1`, in.Username).Scan(&hash); err != nil {
			http.Error(w, "invalid username or password", 401)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
			http.Error(w, "invalid username or password", 401)
			return
		}

		sid, err := api.NewSessionID()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		router.SetSession(sid, in.Username)

		api.WriteJSON(w, 200, map[string]string{
			"token":    sid,
			"username": in.Username,
		})
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "bearer") {
			tok := strings.TrimSpace(parts[1])
			if tok != "" {
				router.DeleteSession(tok)
			}
		}
	}
	w.WriteHeader(204)
}

func AuthMe(w http.ResponseWriter, r *http.Request) {
	user, ok := router.CurrentUser(r)
	if !ok || strings.TrimSpace(user) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	api.WriteJSON(w, http.StatusOK, map[string]string{"username": user})
}

func IsMe(w http.ResponseWriter, r *http.Request) {
	user, ok := router.MustUser(w, r)
	if !ok {
		return
	}

	var in struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	in.Username = strings.TrimSpace(in.Username)
	api.WriteJSON(w, 200, map[string]interface{}{"ok": in.Username == user})
}
