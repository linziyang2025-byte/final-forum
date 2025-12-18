import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../auth";

export default function Header() {
    const { username, logout } = useAuth();
    const nav = useNavigate();

    async function onLogout() {
        try {
            await logout();
            nav("/login");
        } catch {
            nav("/login");
        }
    }

    return (
        <div style={{ padding: 12, borderBottom: "1px solid #ddd" }}>
            <div
                style={{
                    maxWidth: 1000,
                    margin: "0 auto",
                    display: "flex",
                    alignItems: "center",
                    gap: 16,
                }}
            >
                <Link
                    to="/topics"
                    style={{ fontWeight: "bold", marginRight: 16, color: "black", textDecoration: "none" }}
                >
                    Forum
                </Link>

                <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: 10 }}>
                    {username ? (
                        <>
              <span>
                Logged in as <b>{username}</b>
              </span>
                            <button onClick={onLogout}>Logout</button>
                        </>
                    ) : (
                        <Link to="/login">Login</Link>
                    )}
                </div>
            </div>
        </div>
    );
}
