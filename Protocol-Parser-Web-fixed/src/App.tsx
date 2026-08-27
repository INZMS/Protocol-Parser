import { useEffect } from "react";
import { Spin } from "antd";

import LoginPage from "./pages/Login";
import Parser from "./pages/Parser";
import { useAuthStore } from "./store/auth";
import { useSettingsStore } from "./store/settings";


function App(){

  const { initialized, user, initialize } = useAuthStore();
  const settingsInitialized = useSettingsStore((state) => state.initialized);
  const initializeSettings = useSettingsStore((state) => state.initialize);
  const settings = useSettingsStore((state) => state.loginPage);

  useEffect(() => { void initialize(); }, [initialize]);
  useEffect(() => { void initializeSettings(); }, [initializeSettings]);
  useEffect(() => {
    document.title = settings.browserTitleMode === "custom" && settings.browserTitle.trim() ? settings.browserTitle.trim() : settings.systemName;
    let icon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    if (!icon) { icon = document.createElement("link"); icon.rel = "icon"; document.head.appendChild(icon); }
    icon.href = settings.systemIcon || "/favicon.png";
  }, [settings.browserTitleMode, settings.browserTitle, settings.systemIcon, settings.systemName]);

  if (!initialized || !settingsInitialized) return <div className="app-loading"><Spin size="large" /></div>;

  return (

    user ? <Parser /> : <LoginPage />

  )

}


export default App;
