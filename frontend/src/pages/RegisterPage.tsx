import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { api } from "../api";

export default function RegisterPage() {
    const nav = useNavigate();
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [err, setErr] = useState("");

    async function onRegister() {
        setErr("");
        try {
            await api.register(username.trim(), password);
            // 注册后直接去登录
            nav("/login");
        } catch (e: any) {
            setErr(e.message ?? String(e));
        }
    }

    return (
        <div>
            <h2>Register</h2>
            {err && <div style={{ color: "crimson", marginBottom: 12 }}>{err}</div>}

            <div className="card">
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
                        placeholder="Password (min 6 chars)"
                        type="password"
                    />
                </div>

                <button onClick={onRegister} disabled={!username.trim() || password.length < 6}>
                    Register
                </button>

                <div style={{ marginTop: 10 }}>
                    Already have an account? <Link to="/login">Login</Link>
                </div>
            </div>
        </div>
    );
}
