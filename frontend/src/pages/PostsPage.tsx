import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api";
import type { Post } from "../types";

export default function PostsPage() {
    const { topicID } = useParams();
    const tid = Number(topicID);

    const [posts, setPosts] = useState<Post[]>([]);
    const [title, setTitle] = useState("");
    const [content, setContent] = useState("");
    const [err, setErr] = useState("");

    async function refresh() {
        setErr("");
        try {
            const data = await api.listPosts(tid);
            setPosts(Array.isArray(data) ? data : []);
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    useEffect(() => {
        if (tid > 0) refresh();
    }, [tid]);

    async function onCreate() {
        setErr("");
        try {
            await api.createPost(tid, title.trim(), content.trim());
            setTitle("");
            setContent("");
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    async function onDelete(p: Post) {
        if (!confirm(`Delete post "${p.title}"?`)) return;
        setErr("");
        try {
            await api.deletePost(p.id);
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    return (
        <div>
            <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
                <h2 style={{ flex: 1 }}>Posts (Topic #{tid})</h2>
                <Link to="/topics">Back</Link>
            </div>

            <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 12 }}>
                <input
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    placeholder="Post title"
                />
                <textarea
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    placeholder="Post content"
                    rows={4}
                />
                <div style={{ display: "flex", gap: 8 }}>
                    <button onClick={onCreate} disabled={!title.trim() || !content.trim()}>
                        Create
                    </button>
                    <button onClick={refresh}>Refresh</button>
                </div>
            </div>

            {err && <div style={{ color: "crimson" }}>{err}</div>}

            <ul>
                {posts.map((p) => (
                    <li key={p.id} className="card">
                        <Link to={`/posts/${p.id}`} style={{ marginRight: 12 }}>
                            {p.title}
                        </Link>
                        <button onClick={() => onDelete(p)}>Delete</button>
                    </li>
                ))}
            </ul>

            {posts.length === 0 && <div>No posts yet.</div>}
        </div>
    );
}
