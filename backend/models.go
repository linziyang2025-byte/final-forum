package main

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
