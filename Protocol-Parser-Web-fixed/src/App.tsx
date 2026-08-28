import { lazy, Suspense, useEffect } from "react";
import { Spin } from "antd";

import { useAuthStore } from "./store/auth";
import { useSettingsStore } from "./store/settings";

const LoginPage = lazy(() => import("./pages/Login"));
const Parser = lazy(() => import("./pages/Parser"));

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

  return <Suspense fallback={<div className="app-loading"><Spin size="large" /></div>}>
    {user ? <Parser /> : <LoginPage />}
  </Suspense>;

}


export default App;
