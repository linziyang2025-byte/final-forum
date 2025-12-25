package router

import (
	"net/http"
	"strings"
	"sync"
)

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var sessions = map[string]string{}
var sessionsMu sync.RWMutex

func SetSession(sid, user string) {
	sessionsMu.Lock()
	sessions[sid] = strings.TrimSpace(user)
	sessionsMu.Unlock()
}

func DeleteSession(sid string) {
	sessionsMu.Lock()
	delete(sessions, sid)
	sessionsMu.Unlock()
}

func GetSessionUser(sid string) (string, bool) {
	sessionsMu.RLock()
	u, ok := sessions[sid]
	sessionsMu.RUnlock()
	u = strings.TrimSpace(u)
	if !ok || u == "" {
		return "", false
	}
	return u, true
}

func CurrentUser(r *http.Request) (string, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	return GetSessionUser(c.Value)
}

func MustUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	u, ok := CurrentUser(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	return u, true
}

func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := CurrentUser(r); !ok {
			http.Error(w, "unauthorized", 401)
			return
		}
		next.ServeHTTP(w, r)
	})
}
