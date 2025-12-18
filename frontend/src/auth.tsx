import React, { createContext, useContext, useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
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
            setUsername(me.username);
        } catch {
            setUsername(null);
        } finally {
            setChecked(true);
        }
    }

    async function logout() {
        await api.logout();
        await refresh();
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
    const nav = useNavigate();
    const loc = useLocation();

    useEffect(() => {
        if (!checked) return;
        if (!username) {
            nav("/login", { replace: true, state: { from: loc.pathname } });
        }
    }, [checked, username, nav, loc.pathname]);

    if (!checked) return <div className="page">Loading...</div>;
    if (!username) return null; // 正在跳转
    return <>{children}</>;
}
