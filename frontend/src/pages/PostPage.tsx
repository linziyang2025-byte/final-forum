import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api";
import type { Comment, Post } from "../types";

export default function PostPage() {
    const { postID } = useParams();
    const pid = Number(postID);

    const [post, setPost] = useState<Post | null>(null);
    const [comments, setComments] = useState<Comment[]>([]);
    const [err, setErr] = useState("");

    const [editTitle, setEditTitle] = useState("");
    const [editContent, setEditContent] = useState("");

    const [newComment, setNewComment] = useState("");

    async function refresh() {
        setErr("");
        try {
            const p = (await api.getPost(pid)) as Post;
            setPost(p);
            setEditTitle(p.title ?? "");
            setEditContent(p.content ?? "");

            const cs = await api.listComments(pid);
            setComments(Array.isArray(cs) ? cs : []);
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    useEffect(() => {
        if (pid > 0) refresh();
    }, [pid]);

    async function onUpdatePost() {
        if (!post) return;
        setErr("");
        try {
            await api.updatePost(post.id, editTitle.trim(), editContent.trim());
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    async function onCreateComment() {
        setErr("");
        try {
            await api.createComment(pid, newComment.trim());
            setNewComment("");
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    async function onEditComment(c: Comment) {
        const content = prompt("Edit comment:", c.content)?.trim();
        if (!content) return;
        setErr("");
        try {
            await api.updateComment(c.id, content);
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    async function onDeleteComment(c: Comment) {
        if (!confirm("Delete this comment?")) return;
        setErr("");
        try {
            await api.deleteComment(c.id);
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    if (!post) {
        return (
            <div>
                <h2>Post #{pid}</h2>
                {err && <div style={{ color: "crimson" }}>{err}</div>}
                <div>Loading...</div>
            </div>
        );
    }

    return (
        <div>
            <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
                <h2 style={{ flex: 1 }}>{post.title}</h2>
                <Link to={`/topics/${post.topicID}/posts`}>Back</Link>
            </div>

            {err && <div style={{ color: "crimson" }}>{err}</div>}

            <h3>Edit Post</h3>
            <input value={editTitle} onChange={(e) => setEditTitle(e.target.value)} />
            <br />
            <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                rows={6}
                style={{ width: "100%", marginTop: 6 }}
            />
            <br />
            <button onClick={onUpdatePost} disabled={!editTitle.trim() || !editContent.trim()}>
                Update
            </button>

            <h3 style={{ marginTop: 18 }}>Comments</h3>

            <textarea
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder="Write a comment..."
                rows={3}
                style={{ width: "100%" }}
            />
            <br />
            <button onClick={onCreateComment} disabled={!newComment.trim()}>
                Comment
            </button>

            <ul style={{ marginTop: 12 }}>
                {comments.map((c) => (
                    <li key={c.id} className="card">
                        <div style={{ fontWeight: 600, marginBottom: 4 }}>
                            {c.author ?? "Anonymous"}
                        </div>
                        <div>
                            {c.content}
                        </div>
                        <div style={{ marginTop: 8 }}>
                            <button onClick={() => onEditComment(c)}>Edit</button>
                            <button onClick={() => onDeleteComment(c)} style={{ marginLeft: 6 }}>
                                Delete
                            </button>
                        </div>
                    </li>
                ))}
            </ul>

            {comments.length === 0 && <div>No comments yet.</div>}
        </div>
    );
}
