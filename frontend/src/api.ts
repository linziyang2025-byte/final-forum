const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

const TOKEN_KEY = "ff_token";

function getToken(): string {
    return localStorage.getItem(TOKEN_KEY)?.trim() ?? "";
}

function setToken(t: string) {
    const v = (t ?? "").trim();
    if (v) localStorage.setItem(TOKEN_KEY, v);
    else localStorage.removeItem(TOKEN_KEY);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const url = `${API_BASE}${path}`;
    const headers = new Headers(options.headers ?? {});

    if (!headers.has("Content-Type") && options.body !== undefined) {
        headers.set("Content-Type", "application/json");
    }

    const tok = getToken();
    if (tok && !headers.has("Authorization")) {
        headers.set("Authorization", `Bearer ${tok}`);
    }

    const res = await fetch(url, {
        ...options,
        headers,
    });

    if (res.status === 204) return undefined as T;

    const text = await res.text();
    let data: any = null;
    if (text) {
        try {
            data = JSON.parse(text);
        } catch {
            data = text;
        }
    }

    if (res.status === 401) {
        throw new Error("UNAUTHORIZED");
    }

    if (!res.ok) {
        const msg = typeof data === "string" ? data : data?.message ?? text ?? `HTTP ${res.status}`;
        throw new Error(msg);
    }

    return data as T;
}

export const api = {
    me: () => request<{ username: string }>("/auth/me"),

    login: async (username: string, password: string) => {
        const out = await request<{ token: string; username: string }>("/auth/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        });
        setToken(out.token);
    },

    register: (username: string, password: string) =>
        request<void>("/auth/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        }),

    logout: async () => {
        try {
            await request<void>("/auth/logout", { method: "POST" });
        } finally {
            setToken("");
        }
    },

    listTopics: () => request<any[]>("/api/topics"),
    createTopic: (name: string) =>
        request("/api/topics", {
            method: "POST",
            body: JSON.stringify({ name }),
        }),
    updateTopic: (id: number, name: string) =>
        request(`/api/topics/${id}`, {
            method: "PUT",
            body: JSON.stringify({ name }),
        }),
    deleteTopic: (id: number) => request(`/api/topics/${id}`, { method: "DELETE" }),

    listPosts: (topicID: number) => request<any[]>(`/api/topics/${topicID}/posts`),
    createPost: (topicID: number, title: string, content: string) =>
        request(`/api/topics/${topicID}/posts`, {
            method: "POST",
            body: JSON.stringify({ title, content }),
        }),
    getPost: (postID: number) => request(`/api/posts/${postID}`),
    updatePost: (postID: number, title: string, content: string) =>
        request(`/api/posts/${postID}`, {
            method: "PUT",
            body: JSON.stringify({ title, content }),
        }),
    deletePost: (postID: number) => request(`/api/posts/${postID}`, { method: "DELETE" }),

    listComments: (postID: number) => request<any[]>(`/api/posts/${postID}/comments`),
    createComment: (postID: number, content: string) =>
        request(`/api/posts/${postID}/comments`, {
            method: "POST",
            body: JSON.stringify({ content }),
        }),
    updateComment: (commentID: number, content: string) =>
        request(`/api/comments/${commentID}`, {
            method: "PUT",
            body: JSON.stringify({ content }),
        }),
    deleteComment: (commentID: number) => request(`/api/comments/${commentID}`, { method: "DELETE" }),
};
