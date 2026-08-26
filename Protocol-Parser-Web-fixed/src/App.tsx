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

  useEffect(() => { void initialize(); }, [initialize]);
  useEffect(() => { void initializeSettings(); }, [initializeSettings]);

  if (!initialized || !settingsInitialized) return <div className="app-loading"><Spin size="large" /></div>;

  return (

    user ? <Parser /> : <LoginPage />

  )

}


export default App;
