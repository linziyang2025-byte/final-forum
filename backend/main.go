package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

type Topic struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Author    string `json:"author"`
	CreatedAt string `json:"createdAt"`
}

type Post struct {
	ID        int    `json:"id"`
	TopicID   int    `json:"topicID"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	CreatedAt string `json:"createdAt"`
}

type Comment struct {
	ID        int    `json:"id"`
	PostID    int    `json:"postID"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	CreatedAt string `json:"createdAt"`
}

type AuthInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func mustOpenDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func mustApplySchema(db *sql.DB, schema string) {
	b, err := os.ReadFile(schema)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(string(b)); err != nil {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If you use cookie sessions, you MUST allow credentials and cannot use '*'.
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

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Fatal(err)
	}
}

func getID(r *http.Request, key string) (int, error) {
	id_s := chi.URLParam(r, key)
	id, err := strconv.Atoi(id_s)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

//----------------正式开始写topic那些-----------------

func listTopics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, name, author, created_at FROM topics ORDER BY id DESC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		out := []Topic{}
		for rows.Next() {
			var t Topic
			if err := rows.Scan(&t.ID, &t.Name, &t.Author, &t.CreatedAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, t)
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func createTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
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

		res, err := db.Exec(`INSERT INTO topics(name, author) VALUES (?, ?)`, in.Name, user)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id64, _ := res.LastInsertId()
		writeJSON(w, 201, map[string]any{
			"id":     int(id64),
			"name":   in.Name,
			"author": user,
		})
	}
}

func updateTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}

		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		var owner string
		err = db.QueryRow(`SELECT author FROM topics WHERE id = ?`, topicID).Scan(&owner)
		if err == sql.ErrNoRows {
			http.Error(w, "topic not found", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if owner != user {
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
		_, err = db.Exec(`UPDATE topics SET name = ? WHERE id = ?`, in.Name, topicID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, 200, map[string]any{"id": topicID, "name": in.Name})
	}
}

func deleteTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mustUser(w, r)
		if !ok {
			return
		}

		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		var owner string
		err = db.QueryRow(`SELECT author FROM topics WHERE id = ?`, topicID).Scan(&owner)
		if err == sql.ErrNoRows {
			http.Error(w, "topic not found", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if owner != user {
			http.Error(w, "forbidden", 403)
			return
		}

		_, err = db.Exec(`DELETE FROM topics WHERE id = ?`, topicID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(204)
	}
}

// --------------------posts 部分--------------------

func listPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		rows, err := db.Query(`SELECT id, topic_id, title, content, author, created_at
			FROM posts WHERE topic_id = ? ORDER BY id DESC`, topicID)
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
		writeJSON(w, http.StatusOK, out)
	}
}

func createPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		user, ok := mustUser(w, r)
		if !ok {
			return
		}
		user = strings.TrimSpace(user)

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
		if in.Content == "" || in.Title == "" {
			http.Error(w, "title and content required", 400)
			return
		}
		res, err := db.Exec(`INSERT INTO posts(topic_id, title, content, author) VALUES(?, ?, ?, ?)`,
			topicID, in.Title, in.Content, user)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		id64, _ := res.LastInsertId()
		writeJSON(w, 201, map[string]any{"id": int(id64)})
	}
}

func getPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		var p Post
		err = db.QueryRow(`SELECT id, topic_id, title, content, author, created_at FROM posts WHERE id = ?`,
			postID).Scan(&p.ID, &p.TopicID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
		if err == sql.ErrNoRows {
			http.Error(w, "post not found", 404)
			return
		} else if err != nil {
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
			http.Error(w, err.Error(), 400)
			return
		}
		var owner string
		err = db.QueryRow(`SELECT author FROM posts WHERE id = ?`, postID).Scan(&owner)
		if err == sql.ErrNoRows {
			http.Error(w, "post not found", 404)
			return
		}
		if err != nil {
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
		if in.Content == "" || in.Title == "" {
			http.Error(w, "title and content required", 400)
			return
		}
		if _, err := db.Exec(`UPDATE posts SET title = ?, content = ? WHERE id = ?`,
			in.Title, in.Content, postID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, 200, map[string]any{"id": postID})
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
			http.Error(w, err.Error(), 400)
			return
		}
		var owner string
		err = db.QueryRow(`SELECT author FROM posts WHERE id = ?`, postID).Scan(&owner)
		if err == sql.ErrNoRows {
			http.Error(w, "post not found", 404)
			return
		}
		if err != nil {
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

//--------------------------comments部分----------------------------

func listComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		rows, err := db.Query(`SELECT id, post_id, content, author, created_at
			FROM comments WHERE post_id = ? ORDER BY id ASC`, postID)
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
		writeJSON(w, http.StatusOK, out)
	}
}

func createComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		user, ok := mustUser(w, r)
		if !ok {
			return
		}
		user = strings.TrimSpace(user)

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
		res, err := db.Exec(`INSERT INTO comments(post_id, content, author) VALUES(?,?,?)`, postID, in.Content, user)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		id64, _ := res.LastInsertId()
		writeJSON(w, 201, map[string]any{"id": int(id64)})
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
			http.Error(w, err.Error(), 400)
			return
		}
		var owner string
		err = db.QueryRow(`SELECT author FROM comments WHERE id = ?`, commentID).Scan(&owner)
		if err == sql.ErrNoRows {
			http.Error(w, "comment not found", 404)
			return
		}
		if err != nil {
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
		_, err = db.Exec(`UPDATE comments SET content = ? WHERE id = ?`, in.Content, commentID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, 200, map[string]any{"id": commentID})
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
			http.Error(w, err.Error(), 400)
			return
		}
		var owner string
		err = db.QueryRow(`SELECT author FROM comments WHERE id = ?`, commentID).Scan(&owner)
		if err == sql.ErrNoRows {
			http.Error(w, "comment not found", 404)
			return
		}
		if err != nil {
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
		w.WriteHeader(http.StatusNoContent)
	}
}

var sessions = map[string]string{}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in AuthInfo
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

		_, err = db.Exec(`INSERT INTO users(username, password_hash) VALUES (?, ?)`, in.Username, string(hash))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]string{"username": in.Username})
	}
}

func login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in AuthInfo
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		in.Username = strings.TrimSpace(in.Username)

		var hash string
		err := db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, in.Username).Scan(&hash)
		if err != nil {
			http.Error(w, "invalid username or password", 401)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
			http.Error(w, "invalid username or password", 401)
			return
		}

		sid, err := newSessionID()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		sessions[sid] = in.Username

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		_ = json.NewEncoder(w).Encode(map[string]string{"username": in.Username})
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err == nil {
		delete(sessions, c.Value)
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

func authMe(w http.ResponseWriter, r *http.Request) {
	u, ok := currentUser(r)
	if !ok || strings.TrimSpace(u) == "" {
		http.Error(w, "not logged in", http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"username": u})
}

func isMe(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "not logged in", 401)
		return
	}
	u, ok := sessions[c.Value]
	if !ok {
		http.Error(w, "not logged in", 401)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"username": u})
}

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

// ----------------main------------------

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
