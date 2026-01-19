import React, { createContext, useContext, useEffect, useState } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { api } from "./api";

type AuthState = {
    username: string | null;
    checked: boolean;
    refresh: () => Promise<void>;
    logout: () => Promise<void>;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [username, setUsername] = useState<string | null>(null);
    const [checked, setChecked] = useState(false);

    async function refresh() {
        try {
            const me = await api.me();
            const u = (me.username ?? "").trim();
            setUsername(u ? u : null);
        } catch (e: any) {
            if (String(e?.message).includes("UNAUTHORIZED")) {
                localStorage.removeItem("ff_token");
                setUsername(null);
            } else {
                setUsername(null);
            }
        } finally {
            setChecked(true);
        }
    }

    async function logout() {
        try {
            await api.logout();
        } finally {
            await refresh();
        }
    }

    useEffect(() => {
        refresh();
    }, []);

    return (
        <AuthContext.Provider value={{ username, checked, refresh, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
    return ctx;
}

export function RequireAuth({ children }: { children: React.ReactNode }) {
    const { username, checked } = useAuth();
    const loc = useLocation();

    if (!checked) return <div className="page">Loading...</div>;

    if (!username) {
        return <Navigate to="/login" replace state={{ from: loc.pathname }} />;
    }

    return <>{children}</>;
}
