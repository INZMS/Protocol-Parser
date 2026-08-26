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
}

const defaults: LoginPageSettings = {
    layoutType: "split",
    splitImage: "/iot-login-hero-v2.png",
    backgroundImage: "/iot-login-fullscreen-clean.png",
    overlayOpacity: 0,
    animationEnabled: true,
    navigationType: "sidebar"
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
