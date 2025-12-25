package main

import (
	"net/http"
	"strings"
)

func withCORS(next http.Handler) http.Handler {
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

func currentUser(r *http.Request) (string, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	u, ok := sessions[c.Value]
	return u, ok
}

func mustUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	u, ok := currentUser(r)
	if !ok || strings.TrimSpace(u) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	return u, true
}

func requireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := currentUser(r); !ok {
			http.Error(w, "unauthorized", 401)
			return
		}
		next.ServeHTTP(w, r)
	})
}
