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
		if in.Username == "" || len(in.Password) < 6 {
			http.Error(w, "invalid username or password", 400)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if _, err := db.Exec(`INSERT INTO users(username, password_hash) VALUES (?, ?)`, in.Username, string(hash)); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") &&
				strings.Contains(err.Error(), "users.username") {
				http.Error(w, "username already taken", 400)
				return
			}
			http.Error(w, "failed to register", 400)
			return
		}

		w.WriteHeader(201)
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
		if err := db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, in.Username).Scan(&hash); err != nil {
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

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(204)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err == nil {
		router.DeleteSession(c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(204)
}

func AuthMe(w http.ResponseWriter, r *http.Request) {
	u, ok := router.CurrentUser(r)
	if !ok {
		api.WriteJSON(w, 200, map[string]interface{}{"username": ""})
		return
	}
	api.WriteJSON(w, 200, map[string]interface{}{"username": u})
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
