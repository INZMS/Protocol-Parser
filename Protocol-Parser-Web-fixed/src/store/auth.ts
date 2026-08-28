import axios from "axios";
import { create } from "zustand";
import { TOKEN_KEY } from "../api/client";

export interface CurrentUser {
    id: number;
    username: string;
    displayName: string;
    role: string;
    email: string;
    phone: string;
    status: number;
    lastLoginAt?: string | null;
    createdAt: string;
    permissions: string[];
}

export const canAccess = (user: CurrentUser | null, code: string) =>
    !!user && (user.permissions || []).includes(code);

interface AuthStore {
    token: string;
    user: CurrentUser | null;
    initialized: boolean;
    loading: boolean;
    login: (username: string, password: string, captchaId: string, captchaCode: string) => Promise<void>;
    initialize: () => Promise<void>;
    refresh: () => Promise<void>;
    updateProfile: (values: { displayName: string; email: string; phone: string }) => Promise<void>;
    logout: () => Promise<void>;
}

const savedToken = localStorage.getItem(TOKEN_KEY) || "";

export const useAuthStore = create<AuthStore>((set, get) => ({
    token: savedToken,
    user: null,
    initialized: false,
    loading: false,
    login: async (username, password, captchaId, captchaCode) => {
        set({ loading: true });
        try {
            const response = await axios.post("/api/auth/login", { username, password, captchaId, captchaCode });
            const { token, user } = response.data;
            localStorage.setItem(TOKEN_KEY, token);
            set({ token, user, initialized: true, loading: false });
        } catch (error) {
            set({ loading: false });
            throw error;
        }
    },
    initialize: async () => {
        if (get().initialized) return;
        const token = localStorage.getItem(TOKEN_KEY);
        if (!token) { set({ token: "", user: null, initialized: true }); return; }
        try {
            const response = await axios.get("/api/auth/me");
            set({ token, user: response.data.user, initialized: true });
        } catch {
            localStorage.removeItem(TOKEN_KEY);
            set({ token: "", user: null, initialized: true });
        }
    },
    refresh: async () => {
        const token = localStorage.getItem(TOKEN_KEY);
        if (!token) return;
        const response = await axios.get("/api/auth/me");
        set({ token, user: response.data.user, initialized: true });
    },
    updateProfile: async (values) => {
        const response = await axios.put("/api/auth/profile", values);
        set({ user: response.data.user });
    },
    logout: async () => {
        try { await axios.post("/api/auth/logout"); } catch { /* 本地令牌仍需清除 */ }
        localStorage.removeItem(TOKEN_KEY);
        set({ token: "", user: null, initialized: true });
    }
}));

window.addEventListener("auth:unauthorized", () => {
    useAuthStore.setState({ token: "", user: null, initialized: true, loading: false });
});
