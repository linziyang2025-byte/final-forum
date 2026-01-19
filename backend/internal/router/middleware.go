package router

import (
	"net/http"
	"strings"
	"sync"
)

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return "", false
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	scheme := strings.ToLower(strings.TrimSpace(parts[0]))
	tok := strings.TrimSpace(parts[1])
	if tok == "" {
		return "", false
	}
	if scheme != "bearer" {
		return "", false
	}
	return GetSessionUser(tok)
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
