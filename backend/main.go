package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 检测非空header（还得是X-User）
func checkAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" || r.Method == "GET" {
			next.ServeHTTP(w, r)
			return
		}
		au := strings.TrimSpace(r.Header.Get("X-User"))
		if au == "" {
			http.Error(w, "missing X-User header", http.StatusUnauthorized)
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
		rows, err := db.Query(`SELECT id, name, created_at FROM topics ORDER BY id DESC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		out := []Topic{}
		for rows.Next() {
			var t Topic
			if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
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
		res, err := db.Exec(`INSERT INTO topics(name) VALUES (?)`, in.Name)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id64, _ := res.LastInsertId()
		writeJSON(w, 201, map[string]any{
			"id": int(id64), "name": in.Name,
		})
	}
}

func updateTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, err.Error(), 400)
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
		topicID, err := getID(r, "topicID")
		if err != nil {
			http.Error(w, err.Error(), 400)
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
		user := strings.TrimSpace(r.Header.Get("X-User"))
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
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), 400)
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
		postID, err := getID(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), 400)
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
		user := strings.TrimSpace(r.Header.Get("X-User"))

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
		commentID, err := getID(r, "commentID")
		if err != nil {
			http.Error(w, err.Error(), 400)
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
		commentID, err := getID(r, "commentID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if _, err := db.Exec(`DELETE FROM comments WHERE id = ?`, commentID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ----------------main------------------

func main() {
	db := mustOpenDB("./forum.db")
	defer db.Close()
	mustApplySchema(db, "./schema.sql")

	r := chi.NewRouter()
	r.Use(withCORS)
	r.Use(checkAuthor)
	//topic
	r.Get("/api/topics", listTopics(db))
	r.Post("/api/topics", createTopic(db))
	r.Put("/api/topics/{topicID}", updateTopic(db))
	r.Delete("/api/topics/{topicID}", deleteTopic(db))
	//post
	r.Get("/api/topics/{topicID}/posts", listPosts(db))
	r.Post("/api/topics/{topicID}/posts", createPost(db))
	//特定 post
	r.Get("/api/posts/{postID}", getPost(db))
	r.Put("/api/posts/{postID}", updatePost(db))
	r.Delete("/api/posts/{postID}", deletePost(db))
	//comment
	r.Get("/api/posts/{postID}/comments", listComments(db))
	r.Post("/api/posts/{postID}/comments", createComment(db))
	//特定 comment
	r.Put("/api/comments/{commentID}", updateComment(db))
	r.Delete("/api/comments/{commentID}", deleteComment(db))

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
