import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import type { Topic } from "../types";

export default function TopicsPage() {
    const [topics, setTopics] = useState<Topic[]>([]);
    const [name, setName] = useState("");
    const [err, setErr] = useState("");

    async function refresh() {
        setErr("");
        try {
            const data = await api.listTopics();
            setTopics(Array.isArray(data) ? data : []);
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    useEffect(() => {
        refresh();
    }, []);

    async function onCreate() {
        setErr("");
        try {
            await api.createTopic(name.trim());
            setName("");
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    async function onRename(t: Topic) {
        const newName = prompt("New topic name:", t.name)?.trim();
        if (!newName) return;
        setErr("");
        try {
            await api.updateTopic(t.id, newName);
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    async function onDelete(t: Topic) {
        if (!confirm(`Delete topic "${t.name}"?`)) return;
        setErr("");
        try {
            await api.deleteTopic(t.id);
            await refresh();
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    return (
        <div>
            <h2>Topics</h2>

            <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
                <input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="New topic name"
                />
                <button onClick={onCreate} disabled={!name.trim()}>
                    Create
                </button>
                <button onClick={refresh}>Refresh</button>
            </div>

            {err && <div style={{ color: "crimson" }}>{err}</div>}

            <ul>
                {topics.map((t) => (
                    <li key={t.id} className="card">
                        <Link to={`/topics/${t.id}/posts`} style={{ marginRight: 12 }}>
                            {t.name}
                        </Link>
                        <button onClick={() => onRename(t)} style={{ marginRight: 6 }}>
                            Rename
                        </button>
                        <button onClick={() => onDelete(t)}>Delete</button>
                    </li>
                ))}
            </ul>

            {topics.length === 0 && <div>No topics yet.</div>}
        </div>
    );
}
