import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { api } from "../api";
import { useAuth } from "../auth";

export default function RegisterPage() {
    const nav = useNavigate();
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [err, setErr] = useState("");
    const { refresh } = useAuth();

    async function onRegister() {
        setErr("");
        try {
            await api.register(username.trim(), password);
            await api.login(username.trim(), password);
            await refresh();
            nav("/topics");
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    return (
        <div>
            <h2>Register</h2>
            {err && <div style={{ color: "crimson", marginBottom: 12 }}>{err}</div>}

            <form
                className="card"
                onSubmit={(e) => {
                    e.preventDefault();
                    onRegister();
                }}
            >
                <div style={{ marginBottom: 10 }}>
                    <input
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        onInput={(e) => setUsername((e.target as HTMLInputElement).value)}
                        placeholder="Username"
                        autoComplete="username"
                    />
                </div>

                <div style={{ marginBottom: 10 }}>
                    <input
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        onInput={(e) => setPassword((e.target as HTMLInputElement).value)}
                        placeholder="Password"
                        type="password"
                        autoComplete="new-password"
                    />
                </div>

                <button type="submit">Register</button>

                <div style={{ marginTop: 10 }}>
                    Already have an account? <Link to="/login">Login</Link>
                </div>
            </form>
        </div>
    );
}
