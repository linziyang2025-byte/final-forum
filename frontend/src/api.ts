const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

function getUser(): string {
    return localStorage.getItem("x_user")?.trim() ?? "";
}

async function request<T>(
    path: string,
    options: RequestInit = {}
): Promise<T> {
    const url = `${API_BASE}${path}`;
    const headers = new Headers(options.headers ?? {});
    headers.set("Content-Type", "application/json");

    const method = (options.method ?? "GET").toUpperCase();
    if (method !== "GET" && method !== "OPTIONS") {
        const user = getUser();
        if (!user) throw new Error("Please set X-User first");
        headers.set("X-User", user);
    }

    const res = await fetch(url, { ...options, headers });

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

    if (!res.ok) {
        const msg =
            typeof data === "string"
                ? data
                : data?.message ?? text ?? `HTTP ${res.status}`;
        throw new Error(msg);
    }

    return data as T;


    if (!res.ok) {
        throw new Error(data?.message ?? text ?? "Request failed");
    }

    return data as T;
}

export const api = {
    // topics
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
    deleteTopic: (id: number) =>
        request(`/api/topics/${id}`, { method: "DELETE" }),

    // posts
    listPosts: (topicID: number) =>
        request<any[]>(`/api/topics/${topicID}/posts`),
    createPost: (topicID: number, title: string, content: string) =>
        request(`/api/topics/${topicID}/posts`, {
            method: "POST",
            body: JSON.stringify({ title, content }),
        }),
    getPost: (postID: number) =>
        request(`/api/posts/${postID}`),
    updatePost: (postID: number, title: string, content: string) =>
        request(`/api/posts/${postID}`, {
            method: "PUT",
            body: JSON.stringify({ title, content }),
        }),
    deletePost: (postID: number) =>
        request(`/api/posts/${postID}`, { method: "DELETE" }),

    // comments
    listComments: (postID: number) =>
        request<any[]>(`/api/posts/${postID}/comments`),
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
    deleteComment: (commentID: number) =>
        request(`/api/comments/${commentID}`, { method: "DELETE" }),
};
