import axios, { AxiosError } from "axios";

export const TOKEN_KEY = "protocol_parser_token";

let configured = false;

export function configureHttpClient() {
    if (configured) return;
    configured = true;
    axios.defaults.timeout = 30_000;
    axios.interceptors.request.use((config) => {
        const token = localStorage.getItem(TOKEN_KEY);
        if (token) config.headers.Authorization = `Bearer ${token}`;
        return config;
    });
    axios.interceptors.response.use(
        (response) => response,
        (error: AxiosError) => {
            const url = String(error.config?.url || "");
            if (error.response?.status === 401 && !url.includes("/api/auth/login")) {
                localStorage.removeItem(TOKEN_KEY);
                window.dispatchEvent(new CustomEvent("auth:unauthorized"));
            }
            return Promise.reject(error);
        },
    );
}

export function getErrorMessage(error: unknown, fallback = "操作失败") {
    if (axios.isAxiosError(error)) {
        const data = error.response?.data as { error?: string; message?: string } | undefined;
        return data?.error || data?.message || (error.code === "ECONNABORTED" ? "请求超时，请稍后重试" : fallback);
    }
    return error instanceof Error && error.message ? error.message : fallback;
}

export async function withRetry<T>(operation: () => Promise<T>, retries = 1): Promise<T> {
    try {
        return await operation();
    } catch (error) {
        if (retries <= 0 || (axios.isAxiosError(error) && error.response && error.response.status < 500)) throw error;
        await new Promise((resolve) => window.setTimeout(resolve, 250));
        return withRetry(operation, retries - 1);
    }
}
