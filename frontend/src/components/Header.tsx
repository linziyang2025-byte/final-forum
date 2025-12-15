import { useState } from "react";
import { Link } from "react-router-dom";

export default function Header() {
    const [user, setUser] = useState(
        localStorage.getItem("x_user") ?? ""
    );

    return (
        <div style={{ padding: 12, borderBottom: "1px solid #ddd" }}>
            <Link to="/topics" style={{ fontWeight: "bold", marginRight: 16 }}>
                Forum
            </Link>

            <input
                value={user}
                onChange={(e) => setUser(e.target.value)}
                placeholder="Username"
            />
            <button
                onClick={() => localStorage.setItem("x_user", user)}
                style={{ marginLeft: 8 }}
            >
                Save
            </button>
        </div>
    );
}
