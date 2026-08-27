import axios from "axios";
import { create } from "zustand";

export type LoginLayoutType = "split" | "background";
export type NavigationType = "sidebar" | "top";

export interface LoginPageSettings {
    layoutType: LoginLayoutType;
    splitImage: string;
    backgroundImage: string;
    overlayOpacity: number;
    animationEnabled: boolean;
    navigationType: NavigationType;
    systemName: string;
    systemNameEn: string;
    menuShortName: string;
    browserTitleMode: "system" | "menu" | "custom";
    browserTitle: string;
    systemIcon: string;
    footerCopyright: string;
    footerSlogan: string;
    developerName: string;
    developerPhone: string;
    systemVersion: string;
}

const defaults: LoginPageSettings = {
    layoutType: "split",
    splitImage: "/iot-login-hero-v2.png",
    backgroundImage: "/iot-login-fullscreen-clean.png",
    overlayOpacity: 0,
    animationEnabled: true,
    navigationType: "sidebar",
    systemName: "协议解析工具",
    systemNameEn: "Protocol Parser Tool",
    menuShortName: "协议解析工具",
    browserTitleMode: "system",
    browserTitle: "",
    systemIcon: "/favicon.png",
    footerCopyright: "智能风控云平台",
    footerSlogan: "让协议解析更简单高效",
    developerName: "张三科技有限公司",
    developerPhone: "",
    systemVersion: "V1.0.0"
};

interface SettingsStore {
    loginPage: LoginPageSettings;
    initialized: boolean;
    loading: boolean;
    initialize: () => Promise<void>;
    saveLoginPage: (value: LoginPageSettings) => Promise<void>;
}

export const useSettingsStore = create<SettingsStore>((set, get) => ({
    loginPage: defaults,
    initialized: false,
    loading: false,
    initialize: async () => {
        if (get().initialized) return;
        try {
            const response = await axios.get("/api/settings/login-page");
            set({ loginPage: { ...defaults, ...response.data.settings }, initialized: true });
        } catch {
            set({ loginPage: defaults, initialized: true });
        }
    },
    saveLoginPage: async (value) => {
        set({ loading: true });
        try {
            const response = await axios.put("/api/settings/login-page", value);
            set({ loginPage: { ...defaults, ...response.data.settings }, loading: false });
        } catch (error) {
            set({ loading: false });
            throw error;
        }
    }
}));

export const loadUserPreference = async <T,>(key: string): Promise<T | null> => {
    const response = await axios.get(`/api/settings/preferences/${encodeURIComponent(key)}`);
    return response.data?.value ?? null;
};

export const saveUserPreference = async <T,>(key: string, value: T): Promise<void> => {
    await axios.put(`/api/settings/preferences/${encodeURIComponent(key)}`, { value });
};
