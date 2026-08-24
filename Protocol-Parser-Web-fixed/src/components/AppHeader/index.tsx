import { QuestionCircleOutlined, DeleteOutlined, CodeOutlined } from "@ant-design/icons";
import { Modal, message } from "antd";
import axios from "axios";

import { useParserStore } from "../../store/parser";

export default function Header() {
    const bumpHistoryVersion = useParserStore((state) => state.bumpHistoryVersion);

    const clearHistory = () => {
        Modal.confirm({
            title: "清空全部解析记录？",
            content: "清空后无法恢复。",
            okText: "清空",
            okButtonProps: { danger: true },
            cancelText: "取消",
            onOk: async () => {
                try {
                    await axios.delete("/api/parser/history");
                    bumpHistoryVersion();
                    message.success("解析记录已清空");
                } catch (error) {
                    const text = axios.isAxiosError(error)
                        ? String(error.response?.data?.error ?? error.message)
                        : "清空解析记录失败";
                    message.error(text);
                    throw error;
                }
            }
        });
    };
    return (
        <div
            style={{
                height: 64,
                background: "#fff",
                borderBottom: "1px solid #e5e7eb",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "0 clamp(14px, 4vw, 28px)",
                gap: 12
            }}
        >
            {/* 左侧 */}
            <div style={{ display: "flex", alignItems: "center", gap: 10, minWidth: 0 }}>
                <div
                    style={{
                        flex: "0 0 auto",
                        width: 34,
                        height: 34,
                        borderRadius: 8,
                        border: "1px solid #d9d9d9",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        color: "#1677ff",
                        fontSize: 16
                    }}
                >
                    <CodeOutlined />
                </div>

                <div style={{ minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 600, lineHeight: "20px", whiteSpace: "nowrap" }}>
                        协议解析工具
                    </div>
                    <div style={{ fontSize: 11, color: "#8c8c8c", lineHeight: "16px" }}>
                        Protocol Parser Tool
                    </div>
                </div>
            </div>

            {/* 右侧 */}
            <div style={{ display: "flex", alignItems: "center", gap: 20, flex: "0 0 auto", fontSize: 13 }}>
                <div style={{ cursor: "pointer" }} onClick={clearHistory} role="button" tabIndex={0}>
                    <QuestionCircleOutlined />
                    <span style={{ marginLeft: 6 }}>使用说明</span>
                </div>

                <div style={{ cursor: "pointer" }}>
                    <DeleteOutlined />
                    <span style={{ marginLeft: 6 }}>清空记录</span>
                </div>

                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <div
                        style={{
                            width: 28,
                            height: 28,
                            borderRadius: "50%",
                            background: "#e6f4ff",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                            fontSize: 12
                        }}
                    >
                        A
                    </div>
                    admin
                </div>
            </div>
        </div>
    );
}
