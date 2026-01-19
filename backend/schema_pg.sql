CREATE TABLE IF NOT EXISTS topics (
                                      id BIGSERIAL PRIMARY KEY,
                                      name TEXT NOT NULL,
                                      author TEXT NOT NULL,
                                      created_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS posts (
                                     id BIGSERIAL PRIMARY KEY,
                                     topic_id BIGINT NOT NULL,
                                     title TEXT NOT NULL,
                                     content TEXT NOT NULL,
                                     author TEXT NOT NULL,
                                     created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS comments (
                                        id BIGSERIAL PRIMARY KEY,
                                        post_id BIGINT NOT NULL,
                                        content TEXT NOT NULL,
                                        author TEXT NOT NULL,
                                        created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS users (
                                     id BIGSERIAL PRIMARY KEY,
                                     username TEXT UNIQUE NOT NULL,
                                     password_hash TEXT NOT NULL,
                                     created_at TIMESTAMPTZ DEFAULT NOW()
    );
