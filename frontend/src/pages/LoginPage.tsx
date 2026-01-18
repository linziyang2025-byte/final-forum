import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { api } from "../api"
import { useAuth } from "../auth";


export default function LoginPage() {
    const nav = useNavigate();
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [err, setErr] = useState("");
    const { refresh } = useAuth();

    async function onLogin() {
        setErr("");
        try {
            await api.login(username.trim(), password);
            await refresh();
            nav("/topics");
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    return (
        <div>
            <h2>Login</h2>
            {err && <div style={{ color: "crimson", marginBottom: 12 }}>{err}</div>}

            <form className="card" onSubmit={(e) => { e.preventDefault(); onLogin(); }}>
                <div style={{ marginBottom: 10 }}>
                    <input
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        placeholder="Username"
                    />
                </div>

                <div style={{ marginBottom: 10 }}>
                    <input
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="Password"
                        type="password"
                    />
                </div>

                <button type="submit" disabled={!username.trim() || !password}>
                    Login
                </button>

                <div style={{ marginTop: 10 }}>
                    No account? <Link to="/register">Register</Link>
                </div>
            </form>
        </div>
    );
}
