import { useEffect, useState, type CSSProperties } from "react";
import axios from "axios";
import { Button, Form, Input, message } from "antd";
import { CodeOutlined, LockOutlined, ReloadOutlined, SafetyCertificateOutlined, UserOutlined } from "@ant-design/icons";

import { useAuthStore } from "../../store/auth";
import { useSettingsStore } from "../../store/settings";

export default function LoginPage() {
    const login = useAuthStore((state) => state.login);
    const loading = useAuthStore((state) => state.loading);
    const settings = useSettingsStore((state) => state.loginPage);
    const [form] = Form.useForm();
    const [error, setError] = useState("");
    const [captcha, setCaptcha] = useState({ id: "", image: "" });
    const [captchaLoading, setCaptchaLoading] = useState(false);

    const loadCaptcha = async () => {
        setCaptchaLoading(true);
        try {
            const response = await axios.get("/api/auth/captcha");
            setCaptcha(response.data.captcha);
        } finally {
            setCaptchaLoading(false);
        }
    };

    useEffect(() => { void loadCaptcha(); }, []);

    const submit = async (values: { username: string; password: string; captchaCode: string }) => {
        setError("");
        try {
            await login(values.username, values.password, captcha.id, values.captchaCode);
            message.success("登录成功");
        } catch (requestError) {
            const text = axios.isAxiosError(requestError)
                ? String(requestError.response?.data?.error || "登录失败，请稍后重试")
                : "登录失败，请稍后重试";
            setError(text);
            form.setFieldValue("captchaCode", "");
            void loadCaptcha();
        }
    };

    return (
        <div
            className={`login-page login-page-${settings.layoutType}${settings.animationEnabled ? "" : " login-motion-disabled"}`}
            style={{
                "--login-background-image": `url("${settings.layoutType === "split" ? settings.splitImage : settings.backgroundImage}")`,
                "--login-overlay-opacity": settings.overlayOpacity
            } as CSSProperties}
        >
            <div className="login-top-brand">
                <div className="login-top-icon"><CodeOutlined /></div>
                <div>
                    <div className="login-top-title">协议解析工具</div>
                    <div className="login-top-subtitle">Protocol Parser Tool</div>
                </div>
            </div>
            <div className="login-shell">
                <section className="login-visual" aria-label="物联网与车联网连接场景" />
                <main className="login-card">
                <div className="login-form-heading">
                    <h1>欢迎登录</h1>
                    <div className="login-description">请输入账号信息进入系统</div>
                </div>

                <Form form={form} layout="vertical" size="large" onFinish={submit} requiredMark={false}>
                    <Form.Item label="用户名" name="username" rules={[{ required: true, message: "请输入用户名" }]}>
                        <Input prefix={<UserOutlined />} placeholder="请输入用户名" autoComplete="username" />
                    </Form.Item>
                    <Form.Item label="密码" name="password" rules={[{ required: true, message: "请输入密码" }]}>
                        <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
                    </Form.Item>
                    <Form.Item label="验证码" required>
                        <div className="captcha-row">
                            <Form.Item name="captchaCode" noStyle rules={[{ required: true, message: "请输入验证码" }]}>
                                <Input prefix={<SafetyCertificateOutlined />} placeholder="请输入验证码" maxLength={4} autoComplete="off" />
                            </Form.Item>
                            <button className="captcha-image-button" type="button" onClick={() => void loadCaptcha()} title="点击刷新验证码">
                                {captcha.image ? <img src={captcha.image} alt="动态验证码" /> : <span>加载中</span>}
                            </button>
                            <Button className="captcha-refresh" icon={<ReloadOutlined spin={captchaLoading} />} onClick={() => void loadCaptcha()} aria-label="刷新验证码" />
                        </div>
                    </Form.Item>
                    {error && <div className="login-error">{error}</div>}
                    <Button className="login-submit" type="primary" htmlType="submit" block loading={loading}>登 录</Button>
                </Form>
                <div className="login-footer">© 2026 张三科技有限公司</div>
                </main>
            </div>
        </div>
    );
}
